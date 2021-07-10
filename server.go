package lazy

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

func NewServer(cc redis.UniversalClient) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		FetchTimeout: 10 * time.Second,
		JobTimeout:   3 * time.Second,
		DeadJobTTL:   1 * time.Hour,

		cc:         cc,
		queueList:  make([]string, 0, 8),
		handlerMap: make(map[string]HandleFunc, 8),
		ctx:        ctx,
		cancel:     cancel,
	}
}

type Server struct {
	FetchTimeout time.Duration
	JobTimeout   time.Duration
	DeadJobTTL   time.Duration

	cc redis.UniversalClient

	jobLock    sync.RWMutex
	queueList  []string
	handlerMap map[string]HandleFunc

	wg sync.WaitGroup

	ctx        context.Context
	cancel     context.CancelFunc
	cancelOnce sync.Once
}

func (s *Server) SetFetchTimeout(d time.Duration) {
	s.FetchTimeout = d
}

type HandleFunc func(ctx context.Context, jobstr string) error

func (s *Server) Register(queue string, handleFunc HandleFunc) {
	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	s.queueList = append(s.queueList, queue)
	s.handlerMap[queue] = handleFunc
}

func (s *Server) GracefulStop() {
	s.cancelOnce.Do(s.cancel)
}

func (s *Server) Run() error {
	if s == nil {
		return ErrNilServer
	}

	s.jobLock.Lock()
	defer s.jobLock.Unlock()

	if len(s.handlerMap) == 0 {
		return ErrNothingToWork
	}

	queues := make([]string, 0, len(s.handlerMap))
	for queue := range s.handlerMap {
		s.wg.Add(1)
		go s.delayJobScheduler(queue)

		queues = append(queues, queuePrefix+queue)
	}

	defer s.wg.Wait()
	for {
		// Context check
		select {
		default:
		case <-s.ctx.Done():
			return nil
		}

		result, err := s.cc.BRPop(s.ctx, s.FetchTimeout, queues...).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			} else if errors.Is(err, context.Canceled) {
				// Server is closed
				return nil
			} else {
				// Something wrong. Stop the server.
				s.GracefulStop()
				return err
			}
		}

		s.wg.Add(1)
		go s.handleJob(strings.TrimPrefix(result[0], queuePrefix), result[1])
	}
}

func (s *Server) handleJob(queue string, jobstr string) {
	defer s.wg.Done()

	handler := s.handlerMap[queue]

	var ctx context.Context
	if s.JobTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.JobTimeout)
		defer cancel()
	}

	if err := handler(ctx, jobstr); err != nil {
		s.addDeadJob(queue, jobstr, err)
	}
}

func (s *Server) addDeadJob(queue string, jobstr string, reason error) {
	l := log.With().
		Str("queue", queue).
		Str("method", "addDeadJob").
		Str("job", jobstr).
		Logger()

	jb, err := json.Marshal(deadJob{
		Job:    jobstr,
		Reason: reason.Error(),
	})
	if err != nil {
		l.Error().Err(err).Msg("fail to json marshal job")
		s.GracefulStop()
		return
	}

	if err := s.cc.ZAdd(context.Background(), deadPrefix+queue, &redis.Z{
		Score:  float64(time.Now().Add(s.DeadJobTTL).Unix()),
		Member: string(jb),
	}).Err(); err != nil {
		l.Error().Err(err).Msg("fail to add dead job")
		s.GracefulStop()
		return
	}
}

// delayJobScheduler is a scheduler for delayed job.
func (s *Server) delayJobScheduler(queue string) {
	defer s.wg.Done()

	l := log.With().
		Str("queue", queue).
		Str("method", "delayJobScheduler").
		Logger()

	tc := time.NewTicker(time.Second)
	for {
		select {
		case <-s.ctx.Done():
			return
		case now := <-tc.C:
			ctx := context.Background()

			jobs, err := s.cc.ZRangeByScore(ctx, delayPrefix+queue, &redis.ZRangeBy{
				Min: "-inf",
				Max: strconv.FormatInt(now.Unix(), 10),
			}).Result()
			if err != nil {
				l.Error().Err(err).Msg("fail to list delayed job")
				s.GracefulStop()
				return
			}

			for _, job := range jobs {
				rem, err := s.cc.ZRem(ctx, delayPrefix+queue, job).Result()
				if err != nil {
					l.Error().Err(err).Msg("fail to remove delay job")
					continue
				}

				if rem == 0 { // if not removed?
					continue
				}

				if err := s.cc.LPush(context.Background(), queuePrefix+queue, job).Err(); err != nil {
					l.Error().Err(err).
						Str("job", job).
						Msg("!!!job missing!!! fail to schedule job")
					s.GracefulStop()
					return
				}
			}
		}
	}
}

func (s *Server) scheduleDeadJob(queue string, deadjob string) error {
	done, err := s.cc.ZRem(context.Background(), deadPrefix+queue, deadjob).Result()
	if err != nil {
		return err
	}

	if done == 0 {
		return ErrJobMissing
	}

	var j deadJob
	if err := json.Unmarshal([]byte(deadjob), &j); err != nil {
		return err
	}

	jb, err := json.Marshal(j.Job)
	if err != nil {
		return err
	}

	if err := s.cc.LPush(
		context.Background(),
		queuePrefix+queue,
		string(jb),
	).Err(); err != nil {
		return err
	}

	return nil
}
