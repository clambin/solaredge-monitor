package plotter

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"io"
	"math"
)

type ScatterPlotter struct {
	BasePlotter
	Legend Legend
}

type Legend struct {
	Increase int
}

func (s ScatterPlotter) Plot(w io.Writer, measurements repository.Measurements, folded bool) (int64, error) {
	if folded {
		measurements = measurements.Fold()
	}

	data, _, _, zRange := measurementsToXYZs(measurements)

	p := s.preparePlot(folded)

	if err := s.addData(p, data, &zRange); err != nil {
		return 0, err
	}

	if zRange.Bound() {
		s.addLegend(p, &zRange)
	}
	return s.writeImage(w, p)
}

func (s ScatterPlotter) addData(p *plot.Plot, data plotter.XYZs, r *Range) error {
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

func measurementsToXYZs(measurements repository.Measurements) (data plotter.XYZs, xRange, yRange, zRange Range) {
	data = make(plotter.XYZs, len(measurements))
	for index, measurement := range measurements {
		t := measurement.Timestamp.Unix()
		//if fold {
		//	t %= 24 * 3600
		//}
		data[index].X = float64(t)
		xRange.Process(data[index].X)
		data[index].Y = measurement.Intensity
		yRange.Process(data[index].Y)
		data[index].Z = measurement.Power
		zRange.Process(data[index].Z)
	}
	return
}
