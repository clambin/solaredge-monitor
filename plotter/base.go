package plotter

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
)

type BasePlotter struct {
	Title    string
	AxisX    Axis
	AxisY    Axis
	Size     Size
	ColorMap palette.ColorMap
	Fold     bool
}

type Axis struct {
	Label      string
	TimeFormat string
}

type Size struct {
	Width  int
	Height int
}

func (bp BasePlotter) makeBasePlot() *plot.Plot {
	p := plot.New()
	p.Title.Text = bp.Title
	p.Title.Padding = vg.Centimeter
	p.X.Label.Text = bp.AxisX.Label
	//p.X.Tick.Label.XAlign = text.XAlignment(vg.Centimeter)
	//p.X.Tick.Label.Rotation = math.Pi / 2
	if bp.AxisX.TimeFormat != "" {
		p.X.Tick.Marker = plot.TimeTicks{Format: bp.AxisX.TimeFormat}
	}
	p.Y.Label.Text = bp.AxisY.Label
	p.Add(plotter.NewGrid())

	return p
}

// This is the width of the legend, experimentally determined.
const legendWidth = vg.Centimeter

func (bp BasePlotter) createImage(p *plot.Plot) *vgimg.PngCanvas {
	rawImg := vgimg.New(vg.Points(float64(bp.Size.Width)), vg.Points(float64(bp.Size.Height)))
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return &vgimg.PngCanvas{Canvas: rawImg}
}
