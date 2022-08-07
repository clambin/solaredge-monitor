package plotter

import (
	"math"
)

type Range struct {
	Min         float64
	Max         float64
	initialized bool
}

func NewRange(min, max float64) *Range {
	return &Range{
		Min:         min,
		Max:         max,
		initialized: true,
	}
}

func (r *Range) init() {
	if !r.initialized {
		r.Min = math.Inf(1)
		r.Max = math.Inf(-1)
		r.initialized = true
	}
}

func (r *Range) Process(value float64) {
	r.init()
	if value < r.Min {
		r.Min = value
	}
	if value > r.Max {
		r.Max = value
	}
}

func (r *Range) GetIntervals(steps int) []float64 {
	values := make([]float64, steps)
	delta := (r.Max - r.Min) / float64(steps)
	value := r.Min
	for i := 0; i < steps; i++ {
		values[i] = value
		value += delta
	}
	return values
}

func (r *Range) Bound() bool {
	r.init()
	return r.Min != math.Inf(+1) && r.Max != math.Inf(-1)
}
