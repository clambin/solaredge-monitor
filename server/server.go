package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/go-chi/chi/v5"
	middleware2 "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"net/http"
)

type Server struct {
	backend    *Generator
	middleware *middleware.PrometheusMetrics
	server     *http.Server
}

var _ prometheus.Collector = &Server{}

func New(port int, db store.DB) *Server {
	s := Server{
		backend: &Generator{DB: db},
		middleware: middleware.NewPrometheusMetrics(middleware.PrometheusMetricsOptions{
			Namespace:   "solaredge",
			Subsystem:   "monitor",
			Application: "solaredge_monitor",
			MetricsType: middleware.Summary,
			//Buckets:     nil,
		}),
	}

	r := chi.NewRouter()
	r.Use(middleware2.Logger)
	r.Use(s.middleware.Handle)
	r.Get("/report", s.report)
	r.Get("/plot/{type}", s.plot)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	return &s
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		if err := s.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("could not start server", err)
			panic(err)
		}
	}()
	<-ctx.Done()
	_ = s.server.Shutdown(context.Background())
}

func (s *Server) Describe(ch chan<- *prometheus.Desc) {
	s.middleware.Describe(ch)
}

func (s *Server) Collect(ch chan<- prometheus.Metric) {
	s.middleware.Collect(ch)
}
