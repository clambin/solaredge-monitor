package plotter

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

type GriddedPlotter struct {
	BasePlotter
	XResolution int
	YResolution int
	XRange      *Range
	YRange      *Range
	Ranges      []float64
}

func (g GriddedPlotter) preparePlot(measurements []store.Measurement) (*plot.Plot, *Sampler, error) {
	data := Sample(measurements, g.Fold, g.XResolution, g.YResolution, g.XRange, g.YRange)

	rows, cols := data.Dims()
	if rows == 0 || cols == 0 {
		return nil, nil, fmt.Errorf("no data to plot")
	}

	return g.BasePlotter.makeBasePlot(), data, nil
}

func (g GriddedPlotter) addLegend(p *plot.Plot) {
	if len(g.Ranges) == 0 {
		return
	}
	palette := g.ColorMap.Palette(len(g.Ranges))
	thumbs := plotter.PaletteThumbnailers(palette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		val := int(g.Ranges[i])
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}
	p.Legend.XOffs = legendWidth
}
