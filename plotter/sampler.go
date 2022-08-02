package plotter

import (
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/plotter"
)

type Sampler struct {
	Fold     bool
	xRange   *Range
	yRange   *Range
	xSamples int
	ySamples int
	x        []float64
	y        []float64
	z        []float64
}

var _ plotter.GridXYZ = &Sampler{}

func Sample(measurements []store.Measurement, fold bool, xSteps, ySteps int, xRange, yRange *Range) *Sampler {
	s := &Sampler{
		Fold:     fold,
		xSamples: xSteps,
		ySamples: ySteps,
		xRange:   xRange,
		yRange:   yRange,
	}

	if s.xRange == nil {
		s.xRange = s.getRange(measurements, "x")
	}
	if s.yRange == nil {
		s.yRange = s.getRange(measurements, "y")
	}

	s.makeAxes()
	s.process(measurements)

	return s
}

func (s *Sampler) getTimestamp(measurement store.Measurement) float64 {
	t := measurement.Timestamp.Unix()
	if s.Fold {
		t = t % (24 * 3600)
	}
	return float64(t)
}

func (s *Sampler) getRange(measurements []store.Measurement, axis string) *Range {
	minMax := Range{}
	for _, measurement := range measurements {
		var value float64
		switch axis {
		case "x":
			value = s.getTimestamp(measurement)
		case "y":
			value = measurement.Intensity
		}
		minMax.Process(value)
	}
	return &minMax
}

func (s *Sampler) makeAxes() {
	s.x = s.xRange.GetIntervals(s.xSamples)
	s.y = s.yRange.GetIntervals(s.ySamples)
}

func (s *Sampler) process(measurements []store.Measurement) {
	xRange := (s.xRange.Max - s.xRange.Min) / float64(s.xSamples)
	yRange := (s.yRange.Max - s.yRange.Min) / float64(s.ySamples)

	s.z = make([]float64, s.xSamples*s.ySamples)
	zCount := make([]float64, len(s.z))

	for _, measurement := range measurements {
		t := s.getTimestamp(measurement)
		c := getIndex(t, s.xRange.Min, xRange, s.xSamples)
		r := getIndex(measurement.Intensity, s.yRange.Min, yRange, s.ySamples)

		s.z[r*s.xSamples+c] += measurement.Power
		zCount[r*s.xSamples+c]++
	}

	for idx, count := range zCount {
		if count > 0 {
			s.z[idx] /= count
		}
	}
}

func getIndex(value, min, step float64, maxSteps int) int {
	index := int((value - min) / step)
	if index >= maxSteps {
		index = maxSteps - 1
	}
	return index
}

func (s *Sampler) Dims() (c, r int) {
	return len(s.x), len(s.y)
}

func (s *Sampler) Z(c, r int) float64 {
	return s.z[r*s.xSamples+c]
}

func (s *Sampler) X(c int) float64 {
	return s.x[c]
}

func (s *Sampler) Y(r int) float64 {
	return s.y[r]
}
