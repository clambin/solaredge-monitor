package collector_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	db := mockdb.NewDB()
	c := collector.New(50*time.Millisecond, db)
	go c.Run()

	start := time.Now()
	timestamp := start
	delta := 25 * time.Millisecond
	cutOff := start.Add(50 * time.Millisecond)

	for timestamp.Before(cutOff) {
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
	}

	assert.Never(t, func() bool { return db.Rows() > 0 }, 100*time.Millisecond, 10*time.Millisecond)

	c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
	timestamp = timestamp.Add(delta)
	c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}

	assert.Eventually(t, func() bool { return db.Rows() == 1 }, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
		return db.Rows() == 2
	}, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
		timestamp = timestamp.Add(delta)
		return db.Rows() == 3
	}, 500*time.Millisecond, 10*time.Millisecond)

	values, _ := db.GetAll()
	timestamps := make(map[time.Time]bool)

	if assert.Len(t, values, 3) {
		for _, entry := range values {
			timestamps[entry.Timestamp] = true
			assert.Equal(t, 50.0, entry.Power)
			assert.Equal(t, 5.0, entry.Intensity)
		}
		assert.Len(t, timestamps, 3)
	}

	c.Stop <- struct{}{}
	assert.Eventually(t, func() bool { return db.Rows() == 4 }, 500*time.Millisecond, 10*time.Millisecond)
}
