package server

import (
	"context"
	"errors"
	"github.com/clambin/go-common/httpserver"
	"github.com/clambin/solaredge-monitor/store"
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

	var err error
	if s.httpServers["app"], err = httpserver.New(
		httpserver.WithMetrics{Application: "solaredge-monitor"},
		httpserver.WithPort{Port: port},
		httpserver.WithHandlers{Handlers: []httpserver.Handler{
			{Path: "/report", Handler: http.HandlerFunc(s.report)},
			{Path: "/plot/{type}", Handler: http.HandlerFunc(s.plot)},
			{Path: "/", Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/report", http.StatusSeeOther)
			})},
		}},
	); err != nil {
		log.WithError(err).Fatalf("failed to create HTTP server")
	}

	if s.httpServers["metrics"], err = httpserver.New(
		httpserver.WithPort{Port: prometheusPort},
		httpserver.WithPrometheus{},
	); err != nil {
		log.WithError(err).Fatalf("failed to create Prometheus metrics server")
	}
	return s
}

func (s *Server) Run(ctx context.Context) {
	s.httpServers.Start()
	<-ctx.Done()
	s.httpServers.Stop(5 * time.Second)
}

type HTTPServers map[string]*httpserver.Server

func (h HTTPServers) Start() {
	for key, server := range h {
		go func(name string, srv *httpserver.Server) {
			if err := srv.Serve(); !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Fatalf("failed to start %s server", name)

			}
		}(key, server)
	}
}

func (h HTTPServers) Stop(timeout time.Duration) {
	for name, srv := range h {
		if err := srv.Shutdown(timeout); err != nil {
			log.WithError(err).Errorf("failed to shut down %s server", name)
		}
	}
}
