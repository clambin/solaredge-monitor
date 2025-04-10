package web

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/plotters"
	"gonum.org/v1/plot/palette/moreland"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"text/template"
	"time"
)

func ReportHandler(repo Repository, logger *slog.Logger) http.Handler {
	type Data struct {
		Args      string
		PlotTypes []string
		FoldTypes []string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start, end, err := parseReportArguments(r)
		if err != nil {
			http.Error(w, "invalid arguments: "+err.Error(), http.StatusBadRequest)
		}

		if start.IsZero() || end.IsZero() {
			redirectWithDataRange(w, r, repo, logger)
			return
		}

		values := make(url.Values)
		values.Add("start", start.Format(time.RFC3339))
		values.Add("end", end.Format(time.RFC3339))

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

func parseReportArguments(r *http.Request) (start, end time.Time, err error) {
	q := r.URL.Query()
	if start, err = parseTimestamp(q.Get("start")); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start time: %w", err)
	}
	if end, err = parseTimestamp(q.Get("end")); err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end time: %w", err)
	}
	return start, end, nil
}

//go:embed templates/*
var templatesFS embed.FS

var tmpl = template.Must(template.ParseFS(templatesFS, "templates/plot.html"))

func PlotHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start, end, fold, err := parsePlotterArguments(r)
		if err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		if start.IsZero() || end.IsZero() {
			http.Error(w, "start/end cannot be zero", http.StatusBadRequest)
			return
		}

		values := make(url.Values)
		values.Add("start", start.Format(time.RFC3339))
		values.Add("end", end.Format(time.RFC3339))
		values.Add("fold", strconv.FormatBool(fold))

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

func parsePlotterArguments(r *http.Request) (start, end time.Time, fold bool, err error) {
	if start, end, err = parseReportArguments(r); err != nil {
		return time.Time{}, time.Time{}, false, err
	}
	if fold, err = strconv.ParseBool(r.URL.Query().Get("fold")); err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid fold: %w", err)
	}
	return start, end, fold, nil
}

type Repository interface {
	Get(from, to time.Time) (repository.Measurements, error)
	GetDataRange() (time.Time, time.Time, error)
}

var _ Repository = &repository.PostgresDB{}

var DefaultXYZConfig = plotters.XYZConfig{
	Title:   "Report",
	X:       "time",
	XTicker: "2006-01-02\n15:04:05",
	Y:       "solar intensity (%)",
	Width:   800,
	Height:  600,
	// heatmap doesn't use the values in Rangers, just the length.
	// needs to be linear, so that the legend is accurate.
	Ranges: []float64{0, 500, 1000, 1500, 2000, 2500, 3000, 3500, 4000},

	ColorMap: moreland.SmoothBlueRed(),
}

func PlotterHandler(
	repository Repository,
	plotter string,
	logger *slog.Logger,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start, end, fold, err := parsePlotterArguments(r)
		if err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		if start.IsZero() || end.IsZero() {
			http.Error(w, "start/end cannot be zero", http.StatusBadRequest)
			return
		}

		measurements, err := repository.Get(start, end)
		if err != nil {
			logger.Error("failed to get measurements from database", "err", err)
			http.Error(w, fmt.Errorf("database: %w", err).Error(), http.StatusInternalServerError)
			return
		}

		if len(measurements) == 0 {
			http.Error(w, "no data", http.StatusOK)
			return
		}

		config := DefaultXYZConfig
		if fold {
			measurements = measurements.Fold()
			config.XTicker = "15:04:05"
		}

		var buf bytes.Buffer
		switch plotter {
		case "scatter":
			_, err = plotters.XYZScatter(&buf, measurements, config)
		case "heatmap":
			_, err = plotters.XYZHeatmap(&buf, measurements, config, 50, 50)
		}
		if err != nil {
			logger.Error("failed to generate plot", "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("ContentType", "image/png")
		_, _ = io.Copy(w, &buf)
	})
}

func redirectWithDataRange(w http.ResponseWriter, r *http.Request, repo Repository, logger *slog.Logger) {
	start, end, err := repo.GetDataRange()
	if err != nil {
		logger.Error("redirect failed: unable to determine data range", "err", err)
		http.Error(w, "database not available", http.StatusInternalServerError)
		return
	}
	values := url.Values{
		"start": []string{start.Format(time.RFC3339)},
		"end":   []string{end.Format(time.RFC3339)},
	}
	redirectURL := r.URL.Path + "?" + values.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
