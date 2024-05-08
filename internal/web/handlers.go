package web

import (
	"embed"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"gonum.org/v1/plot/vg/vgimg"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

//go:embed templates/*
var html embed.FS

var tmpl = template.Must(template.ParseFS(html, "templates/plot.html"))

func PlotHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		args, err := Parse(r)
		if err != nil {
			logger.Error("failed to determine start/stop parameters", "err", err)
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if args.Stop.IsZero() {
			args.Stop = time.Now()
		}
		values := make(url.Values)
		if args.Fold {
			values.Add("fold", "true")
		}
		values.Add("start", args.Start.Format(time.RFC3339))
		values.Add("stop", args.Stop.Format(time.RFC3339))

		data := struct {
			PlotType string
			Args     string
		}{
			PlotType: r.PathValue("plotType"),
			Args:     values.Encode(),
		}

		//	w.WriteHeader(http.StatusOK)
		if err = tmpl.Execute(w, data); err != nil {
			logger.Error("failed to generate page", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func ReportHandler(logger *slog.Logger) http.Handler {
	type Data struct {
		PlotTypes []string
		FoldTypes []string
		Args      string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		args, err := Parse(r)
		if err != nil {
			logger.Error("failed to determine start/stop parameters", "err", err)
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if args.Stop.IsZero() {
			args.Stop = time.Now()
		}
		values := make(url.Values)
		values.Add("start", args.Start.Format(time.RFC3339))
		values.Add("stop", args.Stop.Format(time.RFC3339))

		reportTemplate := template.Must(template.ParseFS(html, "templates/report.html"))
		data := Data{
			PlotTypes: []string{"scatter", "heatmap"},
			FoldTypes: []string{"false", "true"},
			Args:      values.Encode(),
		}

		//	w.WriteHeader(http.StatusOK)
		if err = reportTemplate.Execute(w, data); err != nil {
			logger.Error("failed to generate page", "err", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

type Repository interface {
	Get(from, to time.Time) (repository.Measurements, error)
}

var _ Repository = &repository.PostgresDB{}

type Plotter interface {
	Plot(repository.Measurements, bool) (*vgimg.PngCanvas, error)
}

var _ Plotter = &plotter.ScatterPlotter{}
var _ Plotter = &plotter.HeatmapPlotter{}
var _ Plotter = &plotter.ContourPlotter{}

func PlotterHandler(
	repository Repository,
	plotter Plotter,
	logger *slog.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		args, err := Parse(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		measurements, err := repository.Get(args.Start, args.Stop)
		if err != nil {
			logger.Error("failed to get measurements from database", "err", err)
			http.Error(w, fmt.Errorf("database: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		img, err := plotter.Plot(measurements, args.Fold)
		if err != nil {
			logger.Error("failed to generate plot", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("ContentType", "image/png")
		_, _ = img.WriteTo(w)
	})
}
