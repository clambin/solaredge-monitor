package plotter

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"io"
)

type ContourPlotter struct {
	GriddedPlotter
}

func (c ContourPlotter) Plot(w io.Writer, measurements repository.Measurements, folded bool) (int64, error) {
	p, data, err := c.preparePlot(measurements, folded)
	if err != nil {
		return 0, err
	}
	c.addContour(p, data)
	c.addLegend(p)
	return c.writeImage(w, p)
}

func (c ContourPlotter) addContour(p *plot.Plot, data plotter.GridXYZ) {
	palette := c.ColorMap.Palette(len(c.Ranges))
	ct := plotter.NewContour(data, c.Ranges, palette)
	p.Add(ct)
}
