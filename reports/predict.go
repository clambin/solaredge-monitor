package reports

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/plot"
	"github.com/clambin/solaredge-monitor/reports/classifier"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/plot/vg/vgimg"
	"time"
)

func (server *Server) Predict(start, stop time.Time) (image []byte, err error) {
	var measurements []store.Measurement
	buf := new(bytes.Buffer)

	if measurements, err = server.db.Get(start, stop); err != nil {
		return nil, err
	}

	options := plot.Options{
		Title:   "Predication",
		AxisX:   plot.Axis{Label: "time", TimeFormat: "15:04:05"},
		AxisY:   plot.Axis{Label: "solar intensity (%)"},
		Legend:  plot.Legend{Increase: 100},
		Size:    plot.Size{Width: 800, Height: 600},
		Contour: plot.Contour{Ranges: []float64{1000, 2000, 3000, 3500}},
	}

	var graph *vgimg.PngCanvas
	if graph, err = plot.ContourPlot(measurementsToPredictedGrid(measurements), options); err == nil {
		_, err = graph.WriteTo(buf)
	}

	return buf.Bytes(), err
}

func measurementsToPredictedGrid(measurements []store.Measurement) (data *plot.GridXYZ) {
	const timeStampRange = 3600
	const intensityRange = 10

	xRange := 24 * 3600 / timeStampRange
	yRange := 100 / int(intensityRange)

	x := make([]float64, 0, xRange)
	for i := 0.0; i < 24*3600; i += timeStampRange {
		x = append(x, i)
	}

	y := make([]float64, 0, yRange)
	for i := 0.0; i < 100; i += intensityRange {
		y = append(y, i)
	}

	// test: eliminate all non-zero output
	filtered := make([]store.Measurement, 0, len(measurements))
	for _, measurement := range measurements {
		if measurement.Intensity > 0.0 {
			filtered = append(filtered, measurement)
		}
	}

	log.Info("learning ...")
	c := classifier.New(100, 1000)
	// c.Learn(measurements)
	c.Learn(filtered)

	log.Info("done")

	input := make([]store.Measurement, 0)
	for _, timestamp := range x {
		for _, intensity := range y {
			hour := int(timestamp) / 3600
			rest := int(timestamp) % 3600
			min := rest / 60
			sec := rest % 60
			input = append(input, store.Measurement{
				Timestamp: time.Date(2021, 1, 1, hour, min, sec, 0, time.UTC),
				Intensity: intensity,
			})
		}
	}

	classified := c.Classify(input)

	z := make([]float64, xRange*yRange)

	for _, measurement := range classified {
		row := (measurement.Timestamp.Hour()*3600 + measurement.Timestamp.Minute()*60 + measurement.Timestamp.Second()) / timeStampRange
		col := int(measurement.Intensity / intensityRange)
		index := row*yRange + col

		z[index] += measurement.Power
	}

	return plot.NewGrid(x, y, z)
}
