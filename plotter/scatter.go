package plotter

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"math"
)

type ScatterPlotter struct {
	BasePlotter
	Legend Legend
}

type Legend struct {
	Increase int
}

var _ Plotter = &ScatterPlotter{}

func (s ScatterPlotter) Plot(measurement []store.Measurement) (*vgimg.PngCanvas, error) {
	dataRange := MeasurementsRange(measurement)
	data := measurementsToXYZs(measurement, s.Fold)

	p := s.makeBasePlot()
	c := s.allocateColors(dataRange)

	if err := s.addScatter(p, c, data, dataRange); err != nil {
		return nil, err
	}

	if dataRange.Bound() {
		s.addLegend(p, c, dataRange)
	}
	return s.createImage(p), nil
}

func (s ScatterPlotter) allocateColors(r *Range) palette.ColorMap {
	colors := s.ColorMap
	colors.SetMin(r.Min)
	colors.SetMax(r.Max)
	return colors
}

func (s ScatterPlotter) addScatter(p *plot.Plot, c palette.ColorMap, data plotter.XYZs, r *Range) error {
	sc, err := plotter.NewScatter(data)
	if err != nil {
		return err
	}

	sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		_, _, z := data.XYZ(i)
		d := (z - r.Min) / (r.Max - r.Min)
		color, _ := c.At(d*(r.Max-r.Min) + r.Min)
		return draw.GlyphStyle{Color: color, Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
	}

	p.Add(sc)
	return nil
}

func (s ScatterPlotter) addLegend(p *plot.Plot, c palette.ColorMap, r *Range) {
	increase := int(math.Max(500, float64(s.Legend.Increase)))
	steps := int(r.Max-r.Min) / increase
	thumbs := plotter.PaletteThumbnailers(c.Palette(steps))
	for i := len(thumbs) - 1; i >= 0; i-- {
		var name string
		switch i {
		case 0:
			name = fmt.Sprintf("%d", int(r.Min))
		case len(thumbs) - 1:
			name = fmt.Sprintf("%d", int(r.Max))
		default:
		}
		p.Legend.Add(name, thumbs[i])
	}

	p.Legend.XOffs = legendWidth
}

func measurementsToXYZs(measurements []store.Measurement, fold bool) (data plotter.XYZs) {
	data = make(plotter.XYZs, len(measurements))
	for index, measurement := range measurements {
		t := measurement.Timestamp.Unix()
		if fold {
			t %= 24 * 3600
		}
		data[index].X = float64(t)
		data[index].Y = measurement.Intensity
		data[index].Z = measurement.Power
	}
	return
}
