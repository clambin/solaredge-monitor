package repository_test

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMeasurements_Fold(t *testing.T) {
	const size = 2 * 365
	measurements := make(repository.Measurements, size)
	timestamp := time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC)
	for i := 0; i < size; i++ {
		measurements[i] = repository.Measurement{Timestamp: timestamp}
		timestamp = timestamp.Add(time.Hour)
	}

	measurements = measurements.Fold()
	assert.Len(t, measurements, size)

	for _, measurement := range measurements {
		assert.Equal(t, 1, measurement.Timestamp.YearDay())
	}
}
