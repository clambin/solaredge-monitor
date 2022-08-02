package plotter

import (
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/vg/vgimg"
)

type Plotter interface {
	Plot(measurement []store.Measurement) (*vgimg.PngCanvas, error)
}

type Options struct {
	Title    string
	AxisX    Axis
	AxisY    Axis
	Legend   Legend
	Size     Size
	ColorMap palette.ColorMap
	Contour  Contour
}

type Axis struct {
	Label      string
	TimeFormat string
}

type Legend struct {
	Increase int
}

type Size struct {
	Width  int
	Height int
}

type Contour struct {
	Ranges []float64
}
