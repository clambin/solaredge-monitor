package collector_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSummary(t *testing.T) {
	s := collector.Summary{}

	assert.Zero(t, s.Get().Timestamp)

	start := time.Now()
	s.Add(collector.Metric{Timestamp: start, Value: 100})
	assert.Equal(t, 100.0, s.Get().Value)

	start = time.Now()
	m := collector.Metric{Timestamp: start, Value: 5}
	s.Add(m)
	s.Add(collector.Metric{Timestamp: m.Timestamp.Add(10 * time.Millisecond), Value: 10})
	s.Add(collector.Metric{Timestamp: m.Timestamp.Add(20 * time.Millisecond), Value: 0})
	s.Add(collector.Metric{Timestamp: m.Timestamp.Add(30 * time.Millisecond), Value: 5})

	assert.Equal(t, 5.0, s.Get().Value)
}
