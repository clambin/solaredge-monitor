package web

import (
	"embed"
	"log/slog"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

func addRoutes(m *http.ServeMux, repo Repository, imageCache *ImageCache, logger *slog.Logger) {
	logger = logger.With("component", "handler")
	m.Handle("GET /report", ReportHandler(repo, logger.With("handler", "report")))
	m.Handle("GET /plot/{plotType}", PlotHandler(logger.With("handler", "plot")))
	for _, plotType := range []string{"scatter", "heatmap"} {
		m.Handle("GET /plotter/"+plotType,
			imageCache.Middleware(plotType, logger.With("cache", plotType))(
				PlotterHandler(repo, plotType, logger.With("handler", plotType)),
			),
		)

	}
	m.Handle("/static/", http.FileServer(http.FS(staticFS)))
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})
}
