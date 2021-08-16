package lazy

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
)

type dashboard struct {
	cc redis.UniversalClient

	prefix string
	queues []string
}

func NewDashboard(cc redis.UniversalClient, pathPrefix string, queues ...string) http.Handler {
	d := &dashboard{
		cc:     cc,
		prefix: pathPrefix,
		queues: queues,
	}

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Route(d.prefix, func(r chi.Router) {
		r.Get("/api/stats", d.ListQueueStats)
		r.Get("/api/stats/{queue}", d.GetQueueStat)
	})

	return r
}

func (d *dashboard) GetQueueStat(w http.ResponseWriter, r *http.Request) {
	queue := chi.URLParam(r, "queue")

	if !d.isQueueKnown(queue) {
		http.NotFound(w, r)
		return
	}

	stat, err := getQueueStats(r.Context(), d.cc, queue)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(stat); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (d *dashboard) ListQueueStats(w http.ResponseWriter, r *http.Request) {
	stats, err := d.listQueueInfo(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (d *dashboard) isQueueKnown(queue string) bool {
	for _, q := range d.queues {
		if q == queue {
			return true
		}
	}
	return false
}

func (d *dashboard) listQueueInfo(ctx context.Context) ([]*Stats, error) {
	stats := make([]*Stats, 0, len(d.queues))

	for _, queue := range d.queues {
		stat, err := getQueueStats(ctx, d.cc, queue)
		if err != nil {
			return nil, err
		}

		stats = append(stats, stat)
	}

	return stats, nil
}
