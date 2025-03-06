package repository

import (
	"gonum.org/v1/plot/plotter"
	"log/slog"
	"time"
)

var _ slog.LogValuer = Measurement{}

type Measurement struct {
	Timestamp time.Time `db:"timestamp"`
	Weather   string    `db:"weather"`
	Power     float64   `db:"power"`
	Intensity float64   `db:"intensity"`
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
	baseDate := time.Date(folded[0].Timestamp.Year(), time.January, 1, 0, 0, 0, 0, folded[0].Timestamp.Location())
	for i := range len(folded) {
		hh, mm, ss := folded[i].Timestamp.Clock()
		folded[i].Timestamp = baseDate.Add(
			time.Duration(hh)*time.Hour +
				time.Duration(mm)*time.Minute +
				time.Duration(ss)*time.Second,
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
