package plotters

import "slices"

// A Sampler collects samples to determine the median, average, min and max.
type Sampler struct {
	values []float64
}

func (m *Sampler) Add(value ...float64) {
	m.values = append(m.values, value...)
}

func (m *Sampler) Len() int {
	return len(m.values)
}

func (m *Sampler) Reset() {
	m.values = m.values[:0]
}

func (m *Sampler) Median() float64 {
	if m.Len() == 0 {
		return 0
	}
	slices.Sort(m.values)
	n := len(m.values)
	if n%2 == 1 {
		return m.values[n/2]
	}
	return (m.values[n/2] + m.values[n/2-1]) / 2
}

func (m *Sampler) Average() float64 {
	if m.Len() == 0 {
		return 0
	}
	var total float64
	for _, value := range m.values {
		total += value
	}
	return total / float64(m.Len())
}

func (m *Sampler) Min() float64 {
	if m.Len() == 0 {
		return 0
	}
	return slices.Min(m.values)
}

func (m *Sampler) Max() float64 {
	if m.Len() == 0 {
		return 0
	}
	return slices.Max(m.values)
}
