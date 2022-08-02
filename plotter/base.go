package plotter

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type BasePlotter struct {
	Options Options
	Fold    bool
}

func (bp BasePlotter) makeBasePlot() *plot.Plot {
	p := plot.New()
	p.Title.Text = bp.Options.Title
	p.Title.Padding = vg.Centimeter
	p.X.Label.Text = bp.Options.AxisX.Label
	if bp.Options.AxisX.TimeFormat != "" {
		p.X.Tick.Marker = plot.TimeTicks{Format: bp.Options.AxisX.TimeFormat}
	}
	p.Y.Label.Text = bp.Options.AxisY.Label
	p.Add(plotter.NewGrid())

	return p
}

// This is the width of the legend, experimentally determined.
const legendWidth = vg.Centimeter

func (bp BasePlotter) saveImage(p *plot.Plot) *vgimg.PngCanvas {
	rawImg := vgimg.New(vg.Points(float64(bp.Options.Size.Width)), vg.Points(float64(bp.Options.Size.Height)))
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return &vgimg.PngCanvas{Canvas: rawImg}
}
