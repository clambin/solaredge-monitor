package collector_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	db := NewDB()
	c := collector.New(50*time.Millisecond, db)
	go c.Run()

	start := time.Now()
	ts := start
	delta := 10 * time.Millisecond

	for start.Add(50*time.Millisecond).Before(ts) {
		c.Intensity <- collector.Metric{Timestamp: ts, Value: 0.0}
		ts = ts.Add(delta)
	}

	assert.Never(t, func() bool { return len(db.Get()) > 0 }, 100*time.Millisecond, 10*time.Millisecond)


	assert.Eventually(t, func() bool {
		c.Intensity <- collector.Metric{Timestamp: ts, Value: 5.0 }
		c.Power <- collector.Metric{Timestamp: ts, Value: 50.0}
		ts = ts.Add(delta)
		return len(db.Get()) > 0
	}, 500*time.Millisecond, 10*time.Millisecond)

	values := db.Get()

	if assert.Len(t, values, 1) {
		for _, entry := range values {
			assert.Equal(t, start, entry.Timestamp)
			assert.Equal(t, 50.0, entry.Power)
			assert.Equal(t, 5.0, entry.Intensity)
		}
	}

	c.Stop <- struct{}{}

	time.Sleep(100*time.Millisecond)

}

func TestCollectorMultiple(t *testing.T) {
	db := NewDB()
	c := collector.New(50*time.Millisecond, db)
	go c.Run()

	start := time.Now()
	end   := start.Add(135*time.Millisecond)
	delta := 10 * time.Millisecond

	for start.Before(end) {
		c.Intensity <- collector.Metric{Timestamp: start, Value: 75.0 }
		c.Power <- collector.Metric{Timestamp: start, Value: 500.0}
		start = start.Add(delta)
	}

	c.Stop <- struct{}{}

	time.Sleep(100*time.Millisecond)

	values := db.Get()

	assert.Len(t, values, 3)

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

