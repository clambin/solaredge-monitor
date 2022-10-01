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
	"time"
)

type Server struct {
	backend     *Generator
	httpServers HTTPServers
}

func New(port, prometheusPort int, db store.DB) (s *Server) {
	s = &Server{
		backend:     &Generator{DB: db},
		httpServers: make(HTTPServers),
	}

	r := mux.NewRouter()
	r.Use(middleware.HTTPMetrics)
	r.HandleFunc("/report", s.report).Methods(http.MethodGet)
	r.HandleFunc("/plot/{type}", s.plot).Methods(http.MethodGet)
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	}).Methods(http.MethodGet)

	s.httpServers["app"] = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())

	s.httpServers["metrics"] = &http.Server{
		Addr:    fmt.Sprintf(":%d", prometheusPort),
		Handler: m,
	}

	return s
}

func (s *Server) Run(ctx context.Context) {
	s.httpServers.Start()
	<-ctx.Done()
	s.httpServers.Stop(5 * time.Second)
}

type HTTPServers map[string]*http.Server

func (h HTTPServers) Start() {
	for key, server := range h {
		go func(name string, srv *http.Server) {
			if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Fatalf("failed to start %s server", name)

			}
		}(key, server)
	}
}

func (h HTTPServers) Stop(timeout time.Duration) {
	for name, srv := range h {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		if err := srv.Shutdown(ctx); err != nil {
			log.WithError(err).Errorf("failed to shut down %s server", name)
		}
		cancel()
	}
}
