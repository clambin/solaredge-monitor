package server

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	port    int
	db      store.DB
	backend *reports.Server
}

func New(port int, backend *reports.Server) *Server {
	return &Server{
		port:    port,
		backend: backend,
	}
}

func (server *Server) Run() {
	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	images := http.StripPrefix("/images/", http.FileServer(http.Dir(server.backend.ImagesDirectory())))
	r.PathPrefix("/images/").Handler(images)
	r.HandleFunc("/overview", server.overview).Methods(http.MethodGet)
	r.HandleFunc("/", server.main).Methods(http.MethodGet)
	address := ":8080"
	if server.port > 0 {
		address = fmt.Sprintf(":%d", server.port)
	}

	err := http.ListenAndServe(address, r)
	log.WithError(err).Fatal("failed to start server")
}

// Prometheus metrics
var (
	httpDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_duration_seconds",
		Help: "API duration of HTTP requests.",
	}, []string{"path"})
)

func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}
