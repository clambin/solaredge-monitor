package server

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/web/cache"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type Server struct {
	port    int
	db      store.DB
	backend *reports.Server
	cache   *cache.Cache
}

func New(port int, imagesDirectory string, backend *reports.Server) *Server {
	if imagesDirectory == "" {
		imagesDirectory, _ = os.MkdirTemp("", "")
	}

	return &Server{
		port:    port,
		backend: backend,
		cache:   cache.New(imagesDirectory, 1*time.Hour, 5*time.Minute),
	}
}

func (server *Server) Run() {
	go server.cache.Run()

	r := mux.NewRouter()
	r.Use(prometheusMiddleware)
	r.Path("/metrics").Handler(promhttp.Handler())
	images := http.StripPrefix("/images/", http.FileServer(http.Dir(server.cache.Directory)))
	r.PathPrefix("/images/").Handler(images)
	r.HandleFunc("/", server.main).Methods(http.MethodGet)
	r.HandleFunc("/report", server.report).Methods(http.MethodGet)

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
