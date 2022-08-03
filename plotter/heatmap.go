package plotter

import (
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
)

type HeatmapPlotter struct {
	GriddedPlotter
}

var _ Plotter = &HeatmapPlotter{}

func (h HeatmapPlotter) Plot(measurements []store.Measurement) (*vgimg.PngCanvas, error) {
	p, data, err := h.preparePlot(measurements)
	if err != nil {
		return nil, err
	}
	h.addHeatmap(p, data)
	h.addLegend(p)
	return h.createImage(p), nil
}

func (h HeatmapPlotter) addHeatmap(p *plot.Plot, data *Sampler) {
	palette := h.ColorMap.Palette(len(h.Ranges))
	ct := plotter.NewHeatMap(data, palette)
	p.Add(ct)
}
