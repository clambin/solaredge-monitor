package plotter

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
)

type ContourPlotter struct {
	GriddedPlotter
}

func (c ContourPlotter) Plot(measurements repository.Measurements, folded bool) (*vgimg.PngCanvas, error) {
	p, data, err := c.preparePlot(measurements, folded)
	if err != nil {
		return nil, err
	}
	c.addContour(p, data)
	c.addLegend(p)
	return c.createImage(p), nil
}

func (c ContourPlotter) addContour(p *plot.Plot, data plotter.GridXYZ) {
	palette := c.ColorMap.Palette(len(c.Ranges))
	ct := plotter.NewContour(data, c.Ranges, palette)
	p.Add(ct)
}
