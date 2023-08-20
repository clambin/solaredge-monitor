package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestScraper(t *testing.T) {
	db := mockdb.NewDB()

	power := newMockPowerGetter(t)
	weather := newMockWeatherGetter(t)

	s := New(power, weather, db, slog.Default(), 100*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- s.Run(ctx)
	}()
	assert.Eventually(t, func() bool { return db.Rows() > 0 }, time.Second, 100*time.Millisecond)

	measurements, _ := db.GetAll()
	for _, entry := range measurements {
		assert.Equal(t, 1000.0, entry.Power)
		assert.Equal(t, 75.0, entry.Intensity)
		assert.Equal(t, "SUNNY", entry.Weather)
	}

	cancel()
	assert.ErrorIs(t, <-errCh, context.Canceled)
}
