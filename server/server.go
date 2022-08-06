package server

import (
	"context"
	"fmt"
	metricsServer "github.com/clambin/go-metrics/server"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	port    int
	backend *Reporter
}

func New(port int, db store.DB) *Server {
	return &Server{
		port:    port,
		backend: &Reporter{DB: db},
	}
}

func (server *Server) Run(ctx context.Context) {
	r := metricsServer.GetRouter()
	r.HandleFunc("/", server.main).Methods(http.MethodGet)
	r.HandleFunc("/report", server.report).Methods(http.MethodGet)
	r.HandleFunc("/plot", server.plot).Methods(http.MethodGet)

	address := ":8080"
	if server.port > 0 {
		address = fmt.Sprintf(":%d", server.port)
	}

	go func() {
		err := http.ListenAndServe(address, r)
		log.WithError(err).Fatal("failed to start server")
	}()

	<-ctx.Done()
}
