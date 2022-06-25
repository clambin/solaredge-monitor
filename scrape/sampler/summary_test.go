package sampler_test

import (
	"github.com/clambin/solaredge-monitor/scrape/sampler"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestSummary(t *testing.T) {
	s := sampler.Summary{}

	result := s.Summarize()
	assert.Zero(t, result.Timestamp)
	assert.True(t, math.IsNaN(result.Value))

	start := time.Now()
	s.Add(sampler.Sample{Timestamp: start, Value: 100})
	assert.Equal(t, 100.0, s.Summarize().Value)

	result = s.Summarize()
	assert.Zero(t, result.Timestamp)
	assert.True(t, math.IsNaN(result.Value))

	start = time.Now()
	s.Add(sampler.Sample{Timestamp: start, Value: 5})
	s.Add(sampler.Sample{Timestamp: start.Add(10 * time.Millisecond), Value: 10})
	s.Add(sampler.Sample{Timestamp: start.Add(20 * time.Millisecond), Value: 0})
	s.Add(sampler.Sample{Timestamp: start.Add(30 * time.Millisecond), Value: 5})

	assert.Equal(t, 5.0, s.Summarize().Value)
}
