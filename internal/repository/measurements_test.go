package repository_test

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
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
		Power:     3000.0,
		Intensity: 80.5,
		Weather:   string(tado.SUN),
	}

	assert.Equal(t, `[power=3000 intensity=80.5 weather=SUN]`, m.LogValue().String())
}
