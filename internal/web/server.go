package web

import (
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/html"
	plotterHandler "github.com/clambin/solaredge-monitor/internal/web/handlers/plotter"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	Router            http.Handler
	PrometheusMetrics middleware.ServerMetrics
}

func NewHTTPServer(repo plotterHandler.Repository, logger *slog.Logger) *HTTPServer {
	s := HTTPServer{
		PrometheusMetrics: middleware.NewDefaultServerSummaryMetrics("solaredge", "monitor", ""),
	}

	mw1 := middleware.RequestLogger(logger, slog.LevelInfo, middleware.DefaultRequestLogFormatter)
	mw2 := middleware.WithServerMetrics(s.PrometheusMetrics)

	m := http.NewServeMux()
	reportsHandler := html.ReportHandler{Logger: logger.With("component", "handler", "handler", "report")}
	m.Handle("GET /report", mw1(mw2(reportsHandler)))

	plotHandler := html.PlotHandler{Logger: logger.With("component", "handler", "handler", "plot")}
	m.Handle("GET /plot/{plotType}", mw1(mw2(plotHandler)))

	scatterHandler := makePlotHandler("scatter", repo, logger)
	m.Handle("GET /plotter/scatter", mw1(mw2(scatterHandler)))

	heatmapHandler := makePlotHandler("heatmap", repo, logger)
	m.Handle("GET /plotter/heatmap", mw1(mw2(heatmapHandler)))

	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/report", http.StatusSeeOther)
	})

	s.Router = m

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
