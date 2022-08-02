package plotter

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
)

type ContourPlotter struct {
	BasePlotter
	Fold   bool
	XSteps int
	YSteps int
	XRange *Range
	YRange *Range
}

var _ Plotter = &ContourPlotter{}

func (c ContourPlotter) Plot(measurement []store.Measurement) (*vgimg.PngCanvas, error) {
	data := Sample(measurement, c.Fold, c.XSteps, c.YSteps, c.XRange, c.YRange)

	rows, cols := data.Dims()
	if rows == 0 || cols == 0 {
		return nil, fmt.Errorf("no data to plot")
	}

	p := c.makeBasePlot()
	c.addContour(p, data)

	if len(c.Options.Contour.Ranges) > 0 {
		c.addLegend(p)
	}

	return c.saveImage(p), nil
}

func (c ContourPlotter) addContour(p *plot.Plot, data plotter.GridXYZ) {
	palette := c.Options.ColorMap.Palette(len(c.Options.Contour.Ranges))
	ct := plotter.NewContour(data, c.Options.Contour.Ranges, palette)
	p.Add(ct)
}

func (c ContourPlotter) addLegend(p *plot.Plot) {
	palette := c.Options.ColorMap.Palette(len(c.Options.Contour.Ranges))
	thumbs := plotter.PaletteThumbnailers(palette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		val := int(c.Options.Contour.Ranges[i])
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}
	p.Legend.XOffs = legendWidth
}
