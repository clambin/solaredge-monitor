package web

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"text/template"
	"time"
)

//go:embed templates/*
var templatesFS embed.FS

var tmpl = template.Must(template.ParseFS(templatesFS, "templates/plot.html"))

func PlotHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: if either arg is blank, look up value from db & redirect
		args, err := parseArguments(r)
		if err != nil {
			logger.Error("failed to determine start/stop parameters", "err", err)
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if args.stop.IsZero() {
			args.stop = time.Now()
		}
		values := make(url.Values)
		if args.fold {
			values.Add("fold", "true")
		}
		values.Add("start", args.start.Format(time.RFC3339))
		values.Add("stop", args.stop.Format(time.RFC3339))

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
		// TODO: if either arg is blank, look up value from db & redirect
		args, err := parseArguments(r)
		if err != nil {
			logger.Error("failed to determine start/stop parameters", "err", err)
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if args.stop.IsZero() {
			args.stop = time.Now()
		}
		values := make(url.Values)
		values.Add("start", args.start.Format(time.RFC3339))
		values.Add("stop", args.stop.Format(time.RFC3339))

		reportTemplate := template.Must(template.ParseFS(templatesFS, "templates/report.html"))
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
	Plot(io.Writer, repository.Measurements, bool) (int64, error)
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
		// TODO: moreland.smoothDiverging.Palette can panic (where there's not enough data?)
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic during plot generation", "err", err)
				response := "failed to generate plot"
				if err2, ok := err.(error); ok {
					response += ": " + err2.Error()
				}
				http.Error(w, response, http.StatusInternalServerError)
			}
		}()

		args, err := parseArguments(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		measurements, err := repository.Get(args.start, args.stop)
		if err != nil {
			logger.Error("failed to get measurements from database", "err", err)
			http.Error(w, fmt.Errorf("database: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		var buf bytes.Buffer
		_, err = plotter.Plot(&buf, measurements, args.fold)
		if err != nil {
			logger.Error("failed to generate plot", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("ContentType", "image/png")
		_, _ = io.Copy(w, bytes.NewReader(buf.Bytes()))
	})
}
