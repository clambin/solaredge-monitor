package plotter

import (
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"io"
)

type BasePlotter struct {
	Title    string
	Size     Size
	ColorMap palette.ColorMap
}

func NewBasePlotter(title string) BasePlotter {
	return BasePlotter{
		Title:    title,
		Size:     Size{Width: width, Height: height},
		ColorMap: moreland.SmoothBlueRed(),
	}
}

type Size struct {
	Width  int
	Height int
}

func (bp BasePlotter) preparePlot(folded bool) *plot.Plot {
	p := plot.New()
	p.Title.Text = bp.Title
	p.Title.Padding = vg.Centimeter
	p.X.Label.Text = "time"
	if folded {
		p.X.Tick.Marker = plot.TimeTicks{Format: "15:04:05"}
	} else {
		p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02\n15:04:05"}
	}
	p.Y.Label.Text = "solar intensity (%)"
	p.Add(plotter.NewGrid())

	return p
}

// This is the width of the legend, experimentally determined.
const legendWidth = vg.Centimeter

func (bp BasePlotter) writeImage(w io.Writer, p *plot.Plot) (int64, error) {
	rawImg := vgimg.New(vg.Points(float64(bp.Size.Width)), vg.Points(float64(bp.Size.Height)))
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return vgimg.PngCanvas{Canvas: rawImg}.WriteTo(w)
}
