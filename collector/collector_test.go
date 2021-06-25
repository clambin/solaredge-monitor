package collector_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	db := NewDB()
	c := collector.New(50*time.Millisecond, db)
	go c.Run()

	start := time.Now()
	timestamp := start
	delta := 25 * time.Millisecond
	cuttOff := start.Add(50 * time.Millisecond)

	for timestamp.Before(cuttOff) {
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
	}

	assert.Never(t, func() bool { return len(db.Get()) > 0 }, 100*time.Millisecond, 10*time.Millisecond)

	c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
	timestamp = timestamp.Add(delta)
	c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}

	assert.Eventually(t, func() bool { return len(db.Get()) == 1 }, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		timestamp = timestamp.Add(delta)
		return len(db.Get()) == 2
	}, 500*time.Millisecond, 10*time.Millisecond)

	assert.Eventually(t, func() bool {
		c.Intensity <- collector.Metric{Timestamp: timestamp, Value: 5.0}
		c.Power <- collector.Metric{Timestamp: timestamp, Value: 50.0}
		timestamp = timestamp.Add(delta)
		return len(db.Get()) == 3
	}, 500*time.Millisecond, 10*time.Millisecond)

	values := db.Get()
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
	assert.Eventually(t, func() bool { return len(db.Get()) == 4 }, 500*time.Millisecond, 10*time.Millisecond)
}

type MockDB struct {
	sync.RWMutex
	content []store.Measurement
}

func NewDB() *MockDB {
	return &MockDB{
		content: make([]store.Measurement, 0),
	}
}

func (db *MockDB) Store(measurement store.Measurement) (err error) {
	db.Lock()
	defer db.Unlock()
	db.content = append(db.content, measurement)
	return
}

func (db *MockDB) Get() []store.Measurement {
	db.RLock()
	defer db.RUnlock()
	return db.content
}
