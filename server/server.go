package server

import (
	"fmt"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

type Server struct {
	http.Handler
	*middleware.PrometheusMetrics
	backend *Generator
}

var _ prometheus.Collector = Server{}

func New(db store.DB) *Server {
	s := Server{
		backend: &Generator{DB: db},
		PrometheusMetrics: middleware.NewPrometheusMetrics(middleware.PrometheusMetricsOptions{
			Namespace:   "solaredge",
			Subsystem:   "monitor",
			Application: "solaredge_monitor",
			MetricsType: middleware.Summary,
			//Buckets:     nil,
		}),
	}

	r := chi.NewRouter()
	//r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Logger(slog.Default()))
	r.Use(s.PrometheusMetrics.Handle)
	r.Get("/report", s.report)
	r.Get("/plot/{type}", s.plot)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})
	s.Handler = r

	return &s
}

func (s *Server) parseTimestamps(req *http.Request) (start, stop time.Time, err error) {
	if start, err = parseTimestamp(req, "start", s.backend.GetFirst); err != nil {
		return
	}
	if stop, err = parseTimestamp(req, "stop", s.backend.GetLast); err != nil {
		return
	}

	if stop.Before(start) {
		err = fmt.Errorf("start time is later than stop time")
	}

	return
}

func parseTimestamp(req *http.Request, field string, dbfunc func() (time.Time, error)) (time.Time, error) {
	arg, ok := req.URL.Query()[field]
	if !ok {
		return dbfunc()
	}

	timestamp, err := time.Parse(time.RFC3339, arg[0])
	if err != nil {
		err = fmt.Errorf("invalid format for '%s': %w", field, err)
	}
	return timestamp, err
}
