package repository

import "time"

type Measurement struct {
	Timestamp time.Time
	Power     float64
	Intensity float64
	Weather   string
}

func (m Measurement) Fold() Measurement {
	return Measurement{
		Timestamp: time.Date(
			2023, time.January, 1,
			m.Timestamp.Hour(), m.Timestamp.Minute(), m.Timestamp.Second(), m.Timestamp.Nanosecond(),
			m.Timestamp.Location(),
		),
		Power:     m.Power,
		Intensity: m.Intensity,
		Weather:   m.Weather,
	}
}

type Measurements []Measurement

func (ms Measurements) Fold() Measurements {
	folded := make(Measurements, len(ms))
	for index, measurement := range ms {
		folded[index] = measurement.Fold()
	}
	return folded
}
