package collector_test

import (
	"context"
	collector2 "github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	db := mockdb.NewDB()
	c := collector2.New(50*time.Millisecond, db)
	ctx, cancel := context.WithCancel(context.Background())
	go c.Run(ctx)

	start := time.Now()
	timestamp := start
	delta := 25 * time.Millisecond
	cutOff := start.Add(50 * time.Millisecond)

	for timestamp.Before(cutOff) {
		c.Intensity <- collector2.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
	}

	assert.Never(t, func() bool { return db.Rows() > 0 }, 100*time.Millisecond, 10*time.Millisecond)

	c.Power <- collector2.Metric{Timestamp: timestamp, Value: 50.0}
	timestamp = timestamp.Add(delta)
	c.Power <- collector2.Metric{Timestamp: timestamp, Value: 50.0}

	assert.Eventually(t, func() bool { return db.Rows() == 1 }, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Power <- collector2.Metric{Timestamp: timestamp, Value: 50.0}
		c.Intensity <- collector2.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
		return db.Rows() == 2
	}, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Intensity <- collector2.Metric{Timestamp: timestamp, Value: 5.0}
		c.Power <- collector2.Metric{Timestamp: timestamp, Value: 50.0}
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

	cancel()
	assert.Eventually(t, func() bool { return db.Rows() == 4 }, 500*time.Millisecond, 10*time.Millisecond)
}
