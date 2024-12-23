package median

import "slices"

type Median struct {
	values []float64
}

func (m *Median) Add(value ...float64) {
	m.values = append(m.values, value...)
}

func (m *Median) Len() int {
	return len(m.values)
}

func (m *Median) Reset() {
	m.values = m.values[:0]
}

func (m *Median) Median() float64 {
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
