package plotter

import (
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/vg/vgimg"
)

type Plotter interface {
	Plot(measurement []store.Measurement) (*vgimg.PngCanvas, error)
}
