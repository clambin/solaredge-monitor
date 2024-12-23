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
		assert.Equal(t, 1, measurement.Timestamp.YearDay()) // all timestamps should be 1 Jan
	}
}

// Before:
// BenchmarkMeasurements_Fold/old-16                    954           1252392 ns/op         1966087 B/op          1 allocs/op
// Current:
// BenchmarkMeasurements_Fold/new-16                   1244            951993 ns/op         1966080 B/op          1 allocs/op
func BenchmarkMeasurements_Fold(b *testing.B) {
	const size = 365 * 24 * 4
	measurements := make(repository.Measurements, size)
	timestamp := time.Date(2024, time.December, 23, 0, 0, 0, 0, time.UTC)
	for i := range size {
		measurements[i] = repository.Measurement{Timestamp: timestamp}
		timestamp = timestamp.Add(15 * time.Minute)
	}
	b.ResetTimer()
	b.Run("old", func(b *testing.B) {
		for range b.N {
			measurements2 := measurements.Fold()
			if len(measurements2) != size {
				b.Fatalf("expected %d measurements, got %d", size, len(measurements2))
			}
		}
	})
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

// TODO: XYZer interface
