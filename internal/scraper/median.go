package scraper

type median struct {
	values []float64
}

func (m *median) add(value float64) {
	m.values = append(m.values, value)
}

func (m *median) len() int {
	return len(m.values)
}

func (m *median) reset() {
	m.values = m.values[:0]
}

func (m *median) median() float64 {
	if m.len() == 0 {
		return 0
	}
	defer m.reset()
	n := len(m.values)
	if n%2 == 1 {
		return m.values[n/2]
	}
	return (m.values[n/2] + m.values[n/2-1]) / 2
}
