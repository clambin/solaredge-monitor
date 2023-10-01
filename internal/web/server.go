package web

import (
	"github.com/clambin/go-common/httpserver/middleware"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/html"
	plotterHandler "github.com/clambin/solaredge-monitor/internal/web/handlers/plotter"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	Router            http.Handler
	PrometheusMetrics *middleware.PrometheusMetrics
}

func NewHTTPServer(repo plotterHandler.Repository, logger *slog.Logger) *HTTPServer {
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

	reportsHandler := html.ReportHandler{Logger: logger.With("component", "handler", "handler", "report")}
	r.Get("/report", reportsHandler.Handle)

	plotHandler := html.PlotHandler{Logger: logger.With("component", "handler", "handler", "plot")}
	r.Get("/plot/{plotType}", plotHandler.Handle)

	scatterHandler := makePlotHandler("scatter", repo, logger)
	r.Get("/plotter/scatter", scatterHandler.Handle)

	heatmapHandler := makePlotHandler("heatmap", repo, logger)
	r.Get("/plotter/heatmap", heatmapHandler.Handle)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})

	s.Router = r

	return &s
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
