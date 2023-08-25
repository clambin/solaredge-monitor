package handlers

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"gonum.org/v1/plot/vg/vgimg"
	"log/slog"
	"net/http"
	"time"
)

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

type PlotHandler struct {
	Repository Repository
	Plotter    Plotter
	Logger     *slog.Logger
}

func (h PlotHandler) Handle(w http.ResponseWriter, req *http.Request) {
	args, err := parseArguments(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	measurements, err := h.Repository.Get(args.start, args.stop)
	if err != nil {
		http.Error(w, fmt.Errorf("database: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	img, err := h.Plotter.Plot(measurements, args.fold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("ContentType", "image/png")
	_, _ = img.WriteTo(w)
}
