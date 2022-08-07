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
	data, _, _, zRange := measurementsToXYZs(measurement, s.Fold)

	p := s.makeBasePlot()

	if err := s.addScatter(p, data, &zRange); err != nil {
		return nil, err
	}

	if zRange.Bound() {
		s.addLegend(p, &zRange)
	}
	return s.createImage(p), nil
}

func (s ScatterPlotter) allocateColors(r *Range) palette.ColorMap {
	colors := s.ColorMap
	colors.SetMin(r.Min)
	colors.SetMax(r.Max)
	return colors
}

func (s ScatterPlotter) addScatter(p *plot.Plot, data plotter.XYZs, r *Range) error {
	sc, err := plotter.NewScatter(data)
	if err != nil {
		return err
	}

	colors := s.ColorMap
	colors.SetMin(r.Min)
	colors.SetMax(r.Max)

	sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		_, _, z := data.XYZ(i)
		color, _ := colors.At(z)
		return draw.GlyphStyle{Color: color, Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
	}

	p.Add(sc)
	return nil
}

func (s ScatterPlotter) addLegend(p *plot.Plot, r *Range) {
	increase := int(math.Max(500, float64(s.Legend.Increase)))
	steps := int(r.Max-r.Min) / increase
	thumbs := plotter.PaletteThumbnailers(s.ColorMap.Palette(steps))
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

func measurementsToXYZs(measurements []store.Measurement, fold bool) (data plotter.XYZs, xRange, yRange, zRange Range) {
	data = make(plotter.XYZs, len(measurements))
	for index, measurement := range measurements {
		t := measurement.Timestamp.Unix()
		if fold {
			t %= 24 * 3600
		}
		data[index].X = float64(t)
		xRange.Process(data[index].X)
		data[index].Y = measurement.Intensity
		yRange.Process(data[index].Y)
		data[index].Z = measurement.Power
		zRange.Process(data[index].Z)
	}
	return
}
