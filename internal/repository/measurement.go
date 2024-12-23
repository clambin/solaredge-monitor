package repository

import (
	"gonum.org/v1/plot/plotter"
	"log/slog"
	"time"
)

var _ slog.LogValuer = Measurement{}

type Measurement struct {
	Timestamp time.Time
	Power     float64
	Intensity float64
	Weather   string
}

func (m Measurement) LogValue() slog.Value {
	return slog.GroupValue(
		//slog.Time("timestamp", m.Timestamp),
		slog.Float64("power", m.Power),
		slog.Float64("intensity", m.Intensity),
		slog.String("weather", m.Weather),
	)
}

var _ plotter.XYZer = Measurements{}

type Measurements []Measurement

func (m Measurements) Fold() Measurements {
	if len(m) == 0 {
		return Measurements{}
	}
	folded := make(Measurements, len(m))
	copy(folded, m)
	baseDate := time.Date(2023, time.January, 1, 0, 0, 0, 0, folded[0].Timestamp.Location())
	for i := range len(folded) {
		folded[i].Timestamp = baseDate.Add(
			time.Duration(folded[i].Timestamp.Hour())*time.Hour +
				time.Duration(folded[i].Timestamp.Minute())*time.Minute +
				time.Duration(folded[i].Timestamp.Second())*time.Second +
				time.Duration(folded[i].Timestamp.Nanosecond()),
		)
	}
	return folded
}

func (m Measurements) Len() int {
	return len(m)
}

func (m Measurements) XYZ(i int) (float64, float64, float64) {
	measurement := m[i]
	return float64(measurement.Timestamp.Unix()), measurement.Intensity, measurement.Power
}

func (m Measurements) XY(i int) (float64, float64) {
	timestamp, intensity, _ := m.XYZ(i)
	return timestamp, intensity
}
