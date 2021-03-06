package mockdb

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	"math"
	"sync"
	"time"
)

type MockDB struct {
	sync.RWMutex
	bad     bool
	content []store.Measurement
}

func NewDB() *MockDB {
	return &MockDB{
		content: make([]store.Measurement, 0),
	}
}

func BadDB() *MockDB {
	return &MockDB{
		bad:     true,
		content: make([]store.Measurement, 0),
	}
}

func (db *MockDB) Store(measurement store.Measurement) (err error) {
	db.Lock()
	defer db.Unlock()

	if db.bad {
		return fmt.Errorf("error accessing database")
	}

	db.content = append(db.content, measurement)
	return
}

func (db *MockDB) Rows() int {
	db.RLock()
	defer db.RUnlock()
	return len(db.content)
}

func (db *MockDB) Get(from, to time.Time) (measurements []store.Measurement, err error) {
	db.RLock()
	defer db.RUnlock()

	if db.bad {
		return measurements, fmt.Errorf("error accessing database")
	}

	for _, entry := range db.content {
		if !entry.Timestamp.Before(from) && !entry.Timestamp.After(to) {
			measurements = append(measurements, entry)
		}
	}

	return
}

func (db *MockDB) GetAll() (measurements []store.Measurement, err error) {
	db.RLock()
	defer db.RUnlock()

	if db.bad {
		return measurements, fmt.Errorf("error accessing database")
	}

	measurements = make([]store.Measurement, len(db.content))
	copy(measurements, db.content)
	return
}

func BuildDB() store.DB {
	db := NewDB()

	start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < 365*24; i++ {
		intensity := 0.0
		hour := start.Hour()
		if hour > 7 && hour < 22 {
			intensity = 100 * math.Sin((float64(hour-7)/15)*math.Pi)
		}

		_ = db.Store(store.Measurement{
			Timestamp: start,
			Power:     intensity * 40,
			Intensity: intensity,
		})
		start = start.Add(1 * time.Hour)
	}

	return db
}
