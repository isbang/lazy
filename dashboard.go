package lazy

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-redis/redis/v8"
)

type dashboard struct {
	cc redis.UniversalClient

	prefix string
	queues []string
}

//go:embed tmpl/*
var tmplFS embed.FS
var tmpl *template.Template

func init() {
	tmpl = template.Must(
		template.New("").
			ParseFS(tmplFS, "tmpl/*.tmpl"),
	)
}

func NewDashboard(cc redis.UniversalClient, pathPrefix string, queues ...string) http.Handler {
	d := &dashboard{
		cc:     cc,
		prefix: path.Clean(pathPrefix),
		queues: queues,
	}

	r := chi.NewRouter()

	r.Use(middleware.Recoverer)

	r.Route(d.prefix, func(r chi.Router) {
		r.Get("/", d.dashboard)
		r.Get("/running", d.running)
		r.Get("/delay", d.delay)
		r.Get("/dead", d.dead)

		r.Get("/api/stats", d.ListQueueStats)
		r.Get("/api/stats/{queue}", d.GetQueueStat)
	})

	return r
}

func (d *dashboard) dashboard(w http.ResponseWriter, r *http.Request) {
	var args struct {
		RootPath string
		Stats    []*Stats
	}

	args.RootPath = d.prefix

	if stats, err := d.listQueueInfo(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		args.Stats = stats
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.tmpl", args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *dashboard) running(w http.ResponseWriter, r *http.Request) {
	var args struct {
		RootPath string
	}

	args.RootPath = d.prefix

	if err := tmpl.ExecuteTemplate(w, "running.tmpl", args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *dashboard) delay(w http.ResponseWriter, r *http.Request) {
	var args struct {
		RootPath string
	}
	args.RootPath = d.prefix

	if err := tmpl.ExecuteTemplate(w, "delay.tmpl", args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (d *dashboard) dead(w http.ResponseWriter, r *http.Request) {
	var args struct {
		RootPath string
	}
	args.RootPath = d.prefix

	if err := tmpl.ExecuteTemplate(w, "dead.tmpl", args); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
