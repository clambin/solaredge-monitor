package plot

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/handlers/arguments"
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

type Handler struct {
	Repository Repository
	Plotter    Plotter
	Logger     *slog.Logger
}

func (h Handler) Handle(w http.ResponseWriter, req *http.Request) {
	args, err := arguments.Parse(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	measurements, err := h.Repository.Get(args.Start, args.Stop)
	if err != nil {
		http.Error(w, fmt.Errorf("database: %w", err).Error(), http.StatusInternalServerError)
		return
	}

	img, err := h.Plotter.Plot(measurements, args.Fold)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("ContentType", "image/png")
	_, _ = img.WriteTo(w)
}
