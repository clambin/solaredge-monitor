package web

import (
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"log/slog"
	"net/http"
)

func addRoutes(m *http.ServeMux, repo Repository, imageCache *ImageCache, logger *slog.Logger) {
	m.Handle("GET /report", ReportHandler(logger.With("component", "handler", "handler", "report")))
	m.Handle("GET /plot/{plotType}", PlotHandler(logger.With("component", "handler", "handler", "plot")))
	m.Handle("GET /plotter/scatter",
		imageCache.Middleware("scatter", logger.With("cache", "scatter"))(
			makePlotterHandler("scatter", repo, logger),
		),
	)
	m.Handle("GET /plotter/heatmap",
		imageCache.Middleware("heatmap", logger.With("cache", "heatmap"))(
			makePlotterHandler("heatmap", repo, logger),
		),
	)
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})
}

func makePlotterHandler(plotType string, repo Repository, logger *slog.Logger) http.Handler {
	return PlotterHandler(repo, makePlotter(plotType), logger.With("component", "handler", "handler", plotType))
}

func makePlotter(plotType string) Plotter {
	var p Plotter
	switch plotType {
	case "scatter":
		p = plotter.ScatterPlotter{
			BasePlotter: plotter.NewBasePlotter("Power output"),
			Legend:      plotter.Legend{Increase: 100},
		}
	case "contour":
		p = plotter.ContourPlotter{
			GriddedPlotter: plotter.NewGriddedPlotter("Power output"),
		}
	case "heatmap":
		p = plotter.HeatmapPlotter{
			GriddedPlotter: plotter.NewGriddedPlotter("Power output"),
		}
	}
	return p
}
