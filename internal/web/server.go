package web

import (
	"github.com/clambin/solaredge-monitor/internal/web/handlers/html"
	plotterHandler "github.com/clambin/solaredge-monitor/internal/web/handlers/plotter"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"log/slog"
	"net/http"
)

func New(repo plotterHandler.Repository, logger *slog.Logger) http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /report", html.ReportHandler{Logger: logger.With("component", "handler", "handler", "report")})
	m.Handle("GET /plot/{plotType}", html.PlotHandler{Logger: logger.With("component", "handler", "handler", "plot")})
	m.Handle("GET /plotter/scatter", makePlotHandler("scatter", repo, logger))
	m.Handle("GET /plotter/heatmap", makePlotHandler("heatmap", repo, logger))
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})
	return m
}

func makePlotHandler(plotType string, repo plotterHandler.Repository, logger *slog.Logger) plotterHandler.Handler {
	return plotterHandler.Handler{
		Repository: repo,
		Plotter:    makePlotter(plotType),
		Logger:     logger.With("component", "handler", "handler", plotType),
	}
}

func makePlotter(plotType string) plotterHandler.Plotter {
	var p plotterHandler.Plotter
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
