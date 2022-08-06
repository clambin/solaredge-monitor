package plotter

import (
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/plotter"
)

type Sampler struct {
	Fold        bool
	xRange      *Range
	yRange      *Range
	xResolution int
	yResolution int
	x           []float64
	y           []float64
	z           []float64
}

var _ plotter.GridXYZ = &Sampler{}

func Sample(measurements []store.Measurement, fold bool, xResolution, yResolution int, xRange, yRange *Range) *Sampler {
	s := &Sampler{
		Fold:        fold,
		xResolution: xResolution,
		yResolution: yResolution,
		xRange:      xRange,
		yRange:      yRange,
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
	s.x = s.xRange.GetIntervals(s.xResolution)
	s.y = s.yRange.GetIntervals(s.yResolution)
}

func (s *Sampler) process(measurements []store.Measurement) {
	xRange := (s.xRange.Max - s.xRange.Min) / float64(s.xResolution)
	yRange := (s.yRange.Max - s.yRange.Min) / float64(s.yResolution)

	s.z = make([]float64, s.xResolution*s.yResolution)
	zCount := make([]float64, len(s.z))

	for _, measurement := range measurements {
		t := s.getTimestamp(measurement)
		c := getIndex(t, s.xRange.Min, xRange, s.xResolution)
		r := getIndex(measurement.Intensity, s.yRange.Min, yRange, s.yResolution)

		s.z[r*s.xResolution+c] += measurement.Power
		zCount[r*s.xResolution+c]++
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
	return s.z[r*s.xResolution+c]
}

func (s *Sampler) X(c int) float64 {
	return s.x[c]
}

func (s *Sampler) Y(r int) float64 {
	return s.y[r]
}
