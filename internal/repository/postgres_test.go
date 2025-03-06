package repository_test

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"testing"
	"time"
)

func TestStore(t *testing.T) {
	c, connString, err := testutils.NewTestPostgresDB(t.Context(), "solaredge", "username", "password")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(c))
	})

	db, err := repository.NewPostgresDB(connString)
	require.NoError(t, err)

	id, err := db.GetWeatherID("SUN")
	require.NoError(t, err)
	assert.Equal(t, 2, id)

	id, err = db.GetWeatherID("SUN")
	require.NoError(t, err)
	assert.Equal(t, 2, id)

	id, err = db.GetWeatherID("CLOUDY")
	require.NoError(t, err)
	assert.Equal(t, 3, id)

	weather, err := db.GetWeather(3)
	require.NoError(t, err)
	assert.Equal(t, "CLOUDY", weather)

	timestamp := time.Date(2021, 7, 4, 12, 0, 0, 0, time.UTC)
	delta := 15 * time.Minute

	first := timestamp

	for i := range 6 {
		err = db.Store(repository.Measurement{
			Timestamp: timestamp,
			Power:     float64(i),
			Intensity: float64(i),
			Weather:   "RAINING",
		})
		require.NoError(t, err)
		timestamp = timestamp.Add(delta)
	}

	var measurements []repository.Measurement
	measurements, err = db.Get(time.Time{}, time.Time{})

	require.NoError(t, err)
	//require.Len(t, measurements, 6)
	assert.Equal(t, first, measurements[0].Timestamp.UTC())
	assert.Equal(t, 0.0, measurements[0].Power)
	assert.Equal(t, 0.0, measurements[0].Intensity)
	assert.Equal(t, "RAINING", measurements[0].Weather)
	assert.Equal(t, timestamp.Add(-delta), measurements[len(measurements)-1].Timestamp.UTC())
	assert.Equal(t, 5.0, measurements[len(measurements)-1].Power)
	assert.Equal(t, 5.0, measurements[len(measurements)-1].Intensity)
	assert.Equal(t, "RAINING", measurements[len(measurements)-1].Weather)

	allCount := len(measurements)

	measurements, err = db.Get(first, timestamp)
	require.NoError(t, err)
	assert.Equal(t, allCount, len(measurements))

	first, last, err := db.GetDataRange()
	require.NoError(t, err)
	assert.NotZero(t, first)
	assert.NotZero(t, last)

	id, err = db.GetWeatherID("RAINING")
	require.NoError(t, err)
	assert.Equal(t, 4, id)
}

func TestNewPostgresDB_ConnectionString(t *testing.T) {
	c, connString, err := testutils.NewTestPostgresDB(t.Context(), "solaredge", "username", "password")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(c))
	})
	t.Log(connString)

	tests := []struct {
		name             string
		connectionString string
		err              assert.ErrorAssertionFunc
	}{
		{"success", connString, assert.NoError},
		{"blank", "", assert.Error},
		{"invalid", "\r\n", assert.Error},
		{"not postgres", "https://example.com", assert.Error},
		{"no db name", "postgres://username:password@localhost:55890", assert.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := repository.NewPostgresDB(tt.connectionString)
			tt.err(t, err)
		})
	}
}
