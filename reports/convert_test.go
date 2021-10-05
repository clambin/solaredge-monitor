package reports

import (
	"github.com/clambin/solaredge-monitor/store"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMeasurementToGrid(t *testing.T) {
	measurements := []store.Measurement{
		{
			Timestamp: time.Date(2021, 10, 5, 0, 0, 0, 0, time.UTC),
			Intensity: 100,
			Power:     2000,
		},
		{
			Timestamp: time.Date(2021, 10, 5, 23, 59, 59, 0, time.UTC),
			Intensity: 100,
			Power:     4000,
		},
	}

	grid := measurementsToGrid(measurements)
	c, r := grid.Dims()
	assert.Equal(t, 24, c)
	assert.Equal(t, 21, r)
	assert.Equal(t, 2000.0, grid.Z(0, r-1))
	assert.Equal(t, 4000.0, grid.Z(c-1, r-1))
}
