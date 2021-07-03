package reports

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/plot"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
	"time"
)

type Server struct {
	db store.DB
}

func New(db store.DB) *Server {
	return &Server{
		db: db,
	}
}

func (server *Server) Summary(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = server.db.Get(start, stop); err != nil {
		return
	}

	options := plot.Options{
		Title:  "Summary",
		AxisX:  plot.Axis{Label: "time", TimeFormat: "15:04:05"},
		AxisY:  plot.Axis{Label: "solar intensity (%)"},
		Legend: plot.Legend{Increase: 100},
		Size:   plot.Size{Width: 800, Height: 600},
	}

	buf := new(bytes.Buffer)
	var img *vgimg.PngCanvas
	if img, err = plot.ScatterPlot(measurementsToPlotData(measurements, true), options); err == nil {
		_, err = img.WriteTo(buf)
	}

	return buf.Bytes(), err
}

func (server *Server) TimeSeries(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = server.db.Get(start, stop); err != nil {
		return
	}

	options := plot.Options{
		Title:  "Time series",
		AxisX:  plot.Axis{Label: "time", TimeFormat: "2006-01-02\n15:04:05"},
		AxisY:  plot.Axis{Label: "solar intensity (%)"},
		Legend: plot.Legend{Increase: 100},
		Size:   plot.Size{Width: 800, Height: 600},
	}

	buf := new(bytes.Buffer)
	var img *vgimg.PngCanvas
	if img, err = plot.ScatterPlot(measurementsToPlotData(measurements, false), options); err == nil {
		_, err = img.WriteTo(buf)
	}

	return buf.Bytes(), err
}

func (server *Server) Classify(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	if measurements, err = server.db.Get(start, stop); err != nil {
		return
	}

	options := plot.Options{
		Title:   "Classification",
		AxisX:   plot.Axis{Label: "time", TimeFormat: "15:04:05"},
		AxisY:   plot.Axis{Label: "solar intensity (%)"},
		Legend:  plot.Legend{Increase: 100},
		Size:    plot.Size{Width: 800, Height: 600},
		Contour: plot.Contour{Ranges: []float64{1000, 2000, 3000, 3500, 3800, 4000}},
	}

	buf := new(bytes.Buffer)
	var graph *vgimg.PngCanvas
	if graph, err = plot.ContourPlot(measurementsToGrid(measurements), options); err == nil {
		_, err = graph.WriteTo(buf)
	}

	return buf.Bytes(), err
}

func measurementsToPlotData(measurements []store.Measurement, fold bool) (data plotter.XYZs) {
	data = make(plotter.XYZs, len(measurements))
	for index, measurement := range measurements {
		if fold {
			data[index].X = float64(measurement.Timestamp.Unix() % (24 * 60 * 60))
		} else {
			data[index].X = float64(measurement.Timestamp.Unix())
		}
		data[index].Y = measurement.Intensity
		data[index].Z = measurement.Power
	}
	return
}

func measurementsToGrid(measurements []store.Measurement) (data *plot.GridXYZ) {
	const timeStampInterval = 3600
	const intensityInterval = 10

	xRange := 24 * 3600 / timeStampInterval
	yRange := 100 / int(intensityInterval)

	x := make([]float64, 0, xRange)
	for i := 0.0; i < 24*3600; i += timeStampInterval {
		x = append(x, i)
	}

	y := make([]float64, 0, yRange)
	for i := 0.0; i < 100; i += intensityInterval {
		y = append(y, i)
	}

	z := make([]float64, xRange*yRange)
	zCounts := make([]int, xRange*yRange)

	for _, measurement := range measurements {
		r := (measurement.Timestamp.Hour()*3600 + measurement.Timestamp.Minute()*60 + measurement.Timestamp.Second()) / timeStampInterval
		c := int(measurement.Intensity / intensityInterval)
		index := r*yRange + c

		z[index] += measurement.Power
		zCounts[index]++
	}

	for index, zCount := range zCounts {
		if zCount != 0 {
			z[index] = z[index] / float64(zCount)
		}
	}

	return plot.NewGrid(x, y, z)
}
