package web

import (
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"log/slog"
	"net/http"
)

func New(repo Repository, logger *slog.Logger) http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /report", ReportHandler(logger.With("component", "handler", "handler", "report")))
	m.Handle("GET /plot/{plotType}", PlotHandler(logger.With("component", "handler", "handler", "plot")))
	m.Handle("GET /plotter/scatter", makePlotHandler("scatter", repo, logger))
	m.Handle("GET /plotter/heatmap", makePlotHandler("heatmap", repo, logger))
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})
	return m
}

func makePlotHandler(plotType string, repo Repository, logger *slog.Logger) http.Handler {
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
