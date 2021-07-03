package feeder_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/feeder"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFeeder(t *testing.T) {
	db := mockdb.NewDB()
	coll := collector.New(15*time.Minute, db)
	ctx, cancel := context.WithCancel(context.Background())
	go coll.Run(ctx)

	var power, intensity []collector.Metric

	timestamp := time.Date(2021, 06, 27, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 30; i++ {
		power = append(power, collector.Metric{
			Timestamp: timestamp,
			Value:     2000.0,
		})
		timestamp = timestamp.Add(30 * time.Second)
		intensity = append(intensity, collector.Metric{
			Timestamp: timestamp,
			Value:     80,
		})
		timestamp = timestamp.Add(30 * time.Second)
	}

	err := feeder.FeedMetrics(power, intensity, coll)
	assert.NoError(t, err)

	intensity = nil
	power = nil

	for i := 0; i < 30; i++ {
		intensity = append(intensity, collector.Metric{
			Timestamp: timestamp,
			Value:     80,
		})
		timestamp = timestamp.Add(30 * time.Second)
		power = append(power, collector.Metric{
			Timestamp: timestamp,
			Value:     2000.0,
		})
		timestamp = timestamp.Add(30 * time.Second)
	}

	err = feeder.FeedMetrics(power, intensity, coll)
	assert.NoError(t, err)

	cancel()
	assert.Eventually(t, func() bool { return db.Rows() == 4 }, 500*time.Millisecond, 50*time.Millisecond)

	content, _ := db.GetAll()
	if assert.Len(t, content, 4) {
		assert.Equal(t, time.Date(2021, 06, 27, 12, 0, 0, 0, time.UTC), content[0].Timestamp)
		assert.Equal(t, time.Date(2021, 06, 27, 12, 15, 30, 0, time.UTC), content[1].Timestamp)
		assert.Equal(t, time.Date(2021, 06, 27, 12, 30, 30, 0, time.UTC), content[2].Timestamp)
		assert.Equal(t, time.Date(2021, 06, 27, 12, 45, 30, 0, time.UTC), content[3].Timestamp)
	}
	for _, entry := range content {
		assert.Equal(t, 2000.0, entry.Power)
		assert.Equal(t, 80.0, entry.Intensity)
	}
}
