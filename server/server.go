package server

import (
	"context"
	"fmt"
	metricsServer "github.com/clambin/go-metrics/server"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	port    int
	backend *Generator
	router  *mux.Router
}

func New(port int, db store.DB) (s *Server) {
	s = &Server{
		port:    port,
		backend: &Generator{DB: db},
		router:  metricsServer.GetRouter(),
	}

	s.router.HandleFunc("/report", s.report).Methods(http.MethodGet)
	s.router.HandleFunc("/plot/{type}", s.plot).Methods(http.MethodGet)
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	}).Methods(http.MethodGet)

	return s
}

func (server *Server) Run(ctx context.Context) {
	address := ":8080"
	if server.port > 0 {
		address = fmt.Sprintf(":%d", server.port)
	}

	go func() {
		err := http.ListenAndServe(address, server.router)
		log.WithError(err).Fatal("failed to start server")
	}()

	<-ctx.Done()
}
