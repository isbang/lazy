package lazy

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Stats struct {
	Queue   string
	Running int64
	Delayed int64
	Dead    int64
}

func getQueueStats(ctx context.Context, cc redis.UniversalClient, name string) (*Stats, error) {
	pipe := cc.Pipeline()
	runningCmd := pipe.LLen(ctx, queuePrefix+name)
	delayedCmd := pipe.ZCard(ctx, delayPrefix+name)
	deadCmd := pipe.ZCard(ctx, deadPrefix+name)
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	stats := &Stats{Queue: name}

	if i64, err := runningCmd.Result(); err != nil {
		return nil, err
	} else {
		stats.Running = i64
	}

	if i64, err := delayedCmd.Result(); err != nil {
		return nil, err
	} else {
		stats.Delayed = i64
	}

	if i64, err := deadCmd.Result(); err != nil {
		return nil, err
	} else {
		stats.Dead = i64
	}

	return stats, nil
}
