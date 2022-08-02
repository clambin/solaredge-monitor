package reports

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/plotter"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/vg/vgimg"
	"time"
)

type Reporter struct {
	db store.DB
}

func New(db store.DB) *Reporter {
	return &Reporter{
		db: db,
	}
}

func (r *Reporter) Summary(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = r.db.Get(start, stop); err != nil {
		return
	}

	p := plotter.ScatterPlotter{
		BasePlotter: plotter.BasePlotter{
			Options: plotter.Options{
				Title:    "Summary",
				AxisX:    plotter.Axis{Label: "time", TimeFormat: "15:04:05"},
				AxisY:    plotter.Axis{Label: "solar intensity (%)"},
				Legend:   plotter.Legend{Increase: 100},
				Size:     plotter.Size{Width: 800, Height: 600},
				ColorMap: moreland.SmoothBlueRed(),
			},
			Fold: true,
		},
	}

	var img *vgimg.PngCanvas
	if img, err = p.Plot(measurements); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	_, err = img.WriteTo(buf)

	return buf.Bytes(), err
}

func (r *Reporter) TimeSeries(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = r.db.Get(start, stop); err != nil {
		return
	}

	p := plotter.ScatterPlotter{
		BasePlotter: plotter.BasePlotter{
			Options: plotter.Options{
				Title:    "Time series",
				AxisX:    plotter.Axis{Label: "time", TimeFormat: "2006-01-02\n15:04:05"},
				AxisY:    plotter.Axis{Label: "solar intensity (%)"},
				Legend:   plotter.Legend{Increase: 100},
				Size:     plotter.Size{Width: 800, Height: 600},
				ColorMap: moreland.SmoothBlueRed(),
			},
			Fold: false,
		},
	}

	var img *vgimg.PngCanvas
	if img, err = p.Plot(measurements); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	_, err = img.WriteTo(buf)

	return buf.Bytes(), err
}

func (r *Reporter) Classify(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = r.db.Get(start, stop); err != nil {
		return
	}

	p := plotter.ContourPlotter{
		BasePlotter: plotter.BasePlotter{
			Options: plotter.Options{
				Title:    "Classification",
				AxisX:    plotter.Axis{Label: "time", TimeFormat: "15:04:05"},
				AxisY:    plotter.Axis{Label: "solar intensity (%)"},
				Legend:   plotter.Legend{Increase: 100},
				Size:     plotter.Size{Width: 800, Height: 600},
				Contour:  plotter.Contour{Ranges: []float64{1000, 2000, 3000, 3500, 3800, 4000}},
				ColorMap: moreland.SmoothBlueRed(),
			},
			Fold: true,
		},
		XSteps: 48,
		YSteps: 20,
		YRange: &plotter.Range{Min: 0, Max: 100},
	}

	var img *vgimg.PngCanvas
	if img, err = p.Plot(measurements); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	_, err = img.WriteTo(buf)

	return buf.Bytes(), err
}
