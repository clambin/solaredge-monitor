package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-metrics/server/middleware"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	backend    *Generator
	appServer  *http.Server
	promServer *http.Server
}

func New(port, prometheusPort int, db store.DB) (s *Server) {
	s = &Server{backend: &Generator{DB: db}}

	r := mux.NewRouter()
	r.Use(middleware.HTTPMetrics)
	r.HandleFunc("/report", s.report).Methods(http.MethodGet)
	r.HandleFunc("/plot/{type}", s.plot).Methods(http.MethodGet)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	}).Methods(http.MethodGet)
	s.appServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	s.promServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", prometheusPort),
		Handler: m,
	}

	return s
}

func (s *Server) Run(ctx context.Context) {
	go func() {
		if err := s.promServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("failed to start prometheus metrics server")
		}
	}()

	go func() {
		if err := s.appServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.WithError(err).Fatal("failed to start application server")
		}
	}()

	<-ctx.Done()
}
