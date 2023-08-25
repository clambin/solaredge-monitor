package plotter

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
)

type HeatmapPlotter struct {
	GriddedPlotter
}

func (h HeatmapPlotter) Plot(measurements repository.Measurements, folded bool) (*vgimg.PngCanvas, error) {
	p, data, err := h.preparePlot(measurements, folded)
	if err != nil {
		return nil, err
	}
	h.addHeatmap(p, data)
	h.addLegend(p)
	return h.createImage(p), err
}

func (h HeatmapPlotter) addHeatmap(p *plot.Plot, data *Sampler) {
	palette := h.ColorMap.Palette(len(h.Ranges))
	ct := plotter.NewHeatMap(data, palette)
	p.Add(ct)
}
