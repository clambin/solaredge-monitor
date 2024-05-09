package repository_test

import (
	"bytes"
	"github.com/clambin/go-common/testutils"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestMeasurements_Fold(t *testing.T) {
	const size = 2 * 365
	measurements := make(repository.Measurements, size)
	timestamp := time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC)
	for i := range size {
		measurements[i] = repository.Measurement{Timestamp: timestamp}
		timestamp = timestamp.Add(time.Hour)
	}

	measurements = measurements.Fold()
	assert.Len(t, measurements, size)

	for _, measurement := range measurements {
		assert.Equal(t, 1, measurement.Timestamp.YearDay())
	}
}

func TestMeasurement_LogValue(t *testing.T) {
	m := repository.Measurement{
		Timestamp: time.Date(2024, time.March, 26, 12, 0, 0, 0, time.UTC),
		Power:     3000,
		Intensity: .8,
		Weather:   "SUNNY",
	}

	var output bytes.Buffer
	l := testutils.NewTextLogger(&output, slog.LevelInfo)
	l.Info("measurement", "measurement", m)

	assert.Equal(t, `level=INFO msg=measurement measurement.power=3000 measurement.intensity=0.8 measurement.weather=SUNNY
`, output.String())
}
