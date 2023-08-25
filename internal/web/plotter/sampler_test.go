package plotter

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var (
	measurements = repository.Measurements{
		{
			Timestamp: time.Date(2022, time.July, 31, 9, 0, 0, 0, time.UTC),
			Intensity: 25,
			Power:     1000,
		},
		{
			Timestamp: time.Date(2022, time.July, 31, 12, 0, 0, 0, time.UTC),
			Intensity: 100,
			Power:     4000,
		},
		{
			Timestamp: time.Date(2022, time.July, 31, 21, 0, 0, 0, time.UTC),
			Intensity: 10,
			Power:     100,
		},
	}
)

func TestSampler(t *testing.T) {
	// No forced ranges.  Output should be:
	//
	// 55%   4000       0
	// 10%   1000     100
	//       9:00   15:00
	//
	s := Sample(measurements, true, 2, 2, nil, nil)
	require.NotNil(t, s)
	r, c := s.Dims()
	assert.Equal(t, 2, r)
	assert.Equal(t, 2, c)
	assert.Equal(t, 9*3600.0, s.X(0))
	assert.Equal(t, 15*3600.0, s.X(1))
	assert.Equal(t, 10.0, s.Y(0))
	assert.Equal(t, 55.0, s.Y(1))
	assert.Equal(t, 1000.0, s.Z(0, 0))
	assert.Equal(t, 100.0, s.Z(1, 0))
	assert.Equal(t, 4000.0, s.Z(0, 1))
	assert.Equal(t, 0.0, s.Z(1, 1))
}

func TestSampler_WithRanges(t *testing.T) {
	// Forced ranges.  Output should be:
	//
	// 50%      0    4000
	//  0%   1000     100
	//       0:00   12:00
	//
	s := Sample(measurements, true, 2, 2, NewRange(0, 24*3600), NewRange(0, 100))
	require.NotNil(t, s)
	r, c := s.Dims()
	assert.Equal(t, 2, r)
	assert.Equal(t, 2, c)
	assert.Equal(t, 0.0, s.X(0))
	assert.Equal(t, 12*3600.0, s.X(1))
	assert.Equal(t, 0.0, s.Y(0))
	assert.Equal(t, 50.0, s.Y(1))
	assert.Equal(t, 1000.0, s.Z(0, 0))
	assert.Equal(t, 100.0, s.Z(1, 0))
	assert.Equal(t, 0.0, s.Z(0, 1))
	assert.Equal(t, 4000.0, s.Z(1, 1))
}
