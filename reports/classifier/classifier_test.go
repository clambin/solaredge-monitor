package classifier_test

import (
	"github.com/clambin/solaredge-monitor/reports/classifier"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestClassifier(t *testing.T) {
	db := mockdb.BuildDB()
	c := classifier.New(100, 1000)
	measurements, _ := db.GetAll()
	c.Learn(measurements[0:256])

	score := c.Score(measurements[256:512])
	assert.Greater(t, score, 0.75)

	classified := c.Classify(measurements[512:])
	assert.Len(t, classified, len(measurements)-512)
}

// TODO: should we make this part of standard testing?
func TestClassifier_Classify(t *testing.T) {
	c := classifier.New(100, 1000)
	measurements, _ := mockdb.BuildDB().GetAll()
	c.Learn(measurements[0:256])

	delta := 0.0
	count := 0.0
	for index, entry := range c.Classify(measurements[512:]) {
		delta += math.Abs(entry.Power - measurements[512+index].Power)
		count++
	}

	assert.Less(t, delta/count, 12.5)
}

func BenchmarkClassifier(b *testing.B) {
	c := classifier.New(100, 1000)
	measurements, _ := mockdb.BuildDB().GetAll()

	classifications := make([]store.Measurement, 0)

	timestamp := time.Date(2021, 6, 28, 0, 0, 0, 0, time.UTC)
	for timestamp.Day() == 28 {
		for intensity := 0.0; intensity <= 100.0; intensity += 5.0 {
			classifications = append(classifications, store.Measurement{
				Timestamp: timestamp,
				Intensity: intensity,
			})
		}
		timestamp = timestamp.Add(15 * time.Minute)
	}

	b.ResetTimer()
	c.Learn(measurements)
	measurements = c.Classify(classifications)
}
