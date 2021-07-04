package store_test

import (
	"github.com/clambin/solaredge-monitor/store"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

func getDBEnv() (values map[string]string, ok bool) {
	values = make(map[string]string, 0)
	envVars := []string{"pg_host", "pg_port", "pg_database", "pg_user", "pg_password"}

	for _, envVar := range envVars {
		value, found := os.LookupEnv(envVar)
		if !found {
			return values, false
		}
		values[envVar] = value
	}

	return values, true
}

func TestStore(t *testing.T) {
	values, ok := getDBEnv()
	if ok == false {
		t.Log("Could not find all DB env variables. Skipping this test")
		return
	}

	port, err := strconv.Atoi(values["pg_port"])
	assert.Nil(t, err)

	db := store.NewPostgresDB(values["pg_host"], port, values["pg_database"], values["pg_user"], values["pg_password"])
	assert.NotNil(t, db)

	timestamp := time.Date(2021, 7, 4, 12, 0, 0, 0, time.UTC)
	delta := 15 * time.Minute

	first := timestamp

	for i := 0; err == nil && i < 6; i++ {
		err = db.Store(store.Measurement{
			Timestamp: timestamp,
			Power:     float64(i),
			Intensity: float64(i),
		})
		assert.NoError(t, err)
		timestamp = timestamp.Add(delta)
	}

	var measurements []store.Measurement
	measurements, err = db.GetAll()

	if assert.NoError(t, err) && assert.NotEmpty(t, measurements) {
		assert.Equal(t, first, measurements[0].Timestamp)
		assert.Equal(t, 0.0, measurements[0].Power)
		assert.Equal(t, 0.0, measurements[0].Intensity)
		assert.Equal(t, timestamp.Add(-delta), measurements[len(measurements)-1].Timestamp)
		assert.Equal(t, 5.0, measurements[len(measurements)-1].Power)
		assert.Equal(t, 5.0, measurements[len(measurements)-1].Intensity)
	}

	allCount := len(measurements)

	measurements, err = db.Get(first, timestamp)
	assert.NoError(t, err)
	assert.Equal(t, allCount, len(measurements))
}
