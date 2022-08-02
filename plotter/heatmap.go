package plotter

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
)

type HeatmapPlotter struct {
	BasePlotter
	Fold   bool
	XSteps int
	YSteps int
	XRange *Range
	YRange *Range
}

var _ Plotter = &HeatmapPlotter{}

func (h HeatmapPlotter) Plot(measurement []store.Measurement) (*vgimg.PngCanvas, error) {
	data := Sample(measurement, h.Fold, h.XSteps, h.YSteps, h.XRange, h.YRange)

	rows, cols := data.Dims()
	if rows == 0 || cols == 0 {
		return nil, fmt.Errorf("no data to plot")
	}

	p := h.makeBasePlot()
	h.addHeatmap(p, data)

	if len(h.Options.Contour.Ranges) > 0 {
		h.addLegend(p)
	}

	return h.saveImage(p), nil
}

func (h HeatmapPlotter) addHeatmap(p *plot.Plot, data *Sampler) {
	palette := h.Options.ColorMap.Palette(len(h.Options.Contour.Ranges))
	ct := plotter.NewHeatMap(data, palette)
	p.Add(ct)
}

func (h HeatmapPlotter) addLegend(p *plot.Plot) {
	palette := h.Options.ColorMap.Palette(len(h.Options.Contour.Ranges))
	thumbs := plotter.PaletteThumbnailers(palette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		val := int(h.Options.Contour.Ranges[i])
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}
	p.Legend.XOffs = legendWidth
}
