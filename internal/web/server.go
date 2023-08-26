package web

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/plot"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/report"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	Router            http.Handler
	PrometheusMetrics *middleware.PrometheusMetrics
}

func NewHTTPServer(repo plot.Repository, logger *slog.Logger) *HTTPServer {
	s := HTTPServer{
		PrometheusMetrics: middleware.NewPrometheusMetrics(middleware.PrometheusMetricsOptions{
			Namespace:   "solaredge",
			Subsystem:   "monitor",
			Application: "solaredge_monitor",
			MetricsType: middleware.Summary,
		}),
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger, slog.LevelInfo, middleware.DefaultRequestLogFormatter))
	r.Use(s.PrometheusMetrics.Handle)

	reportsHandler := report.ReportsHandler{Logger: logger.With("component", "handler", "hander", "report")}
	r.Get("/report", reportsHandler.Handle)

	scatterHandler := makePlotHandler("scatter", repo, logger)
	r.Get("/plot/scatter", scatterHandler.Handle)

	heatmapHandler := makePlotHandler("heatmap", repo, logger)
	r.Get("/plot/heatmap", heatmapHandler.Handle)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})

	s.Router = r

	return &s
}

func makePlotHandler(plotType string, repo plot.Repository, logger *slog.Logger) *plot.PlotHandler {
	return &plot.PlotHandler{
		Repository: repo,
		Plotter:    makePlotter(plotType),
		Logger:     logger.With("component", "handler", "handler", plotType),
	}
}

func makePlotter(plotType string) plot.Plotter {
	var p plot.Plotter
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
