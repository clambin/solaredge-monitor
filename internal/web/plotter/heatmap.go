package plotter

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"io"
)

type HeatmapPlotter struct {
	GriddedPlotter
}

func (h HeatmapPlotter) Plot(w io.Writer, measurements repository.Measurements, folded bool) (int64, error) {
	p, data, err := h.preparePlot(measurements, folded)
	if err != nil {
		return 0, err
	}
	h.addHeatmap(p, data)
	h.addLegend(p)
	return h.writeImage(w, p)
}

func (h HeatmapPlotter) addHeatmap(p *plot.Plot, data *Sampler) {
	palette := h.ColorMap.Palette(len(h.Ranges))
	ct := plotter.NewHeatMap(data, palette)
	p.Add(ct)
}
