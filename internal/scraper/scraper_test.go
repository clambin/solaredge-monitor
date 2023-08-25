package scraper

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

func TestScraper(t *testing.T) {
	var ran atomic.Bool
	db := mocks.NewRepository(t)
	db.EXPECT().Store(mock.Anything).RunAndReturn(func(measurement repository.Measurement) error {
		if measurement.Power != 1000 {
			return fmt.Errorf("bad power reading: %f", measurement.Power)
		}
		if measurement.Intensity != 75 {
			return fmt.Errorf("bad intensity reading: %f", measurement.Intensity)
		}
		if measurement.Weather != "SUNNY" {
			return fmt.Errorf("bad weather reading: %s", measurement.Weather)
		}
		ran.Store(true)
		return nil
	})

	powerGetter := newMockPowerGetter(t)
	weatherGetter := newMockWeatherGetter(t)

	s := New(powerGetter, weatherGetter, db, slog.Default().With("component", "collector"), 10*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- s.Run(ctx) }()

	assert.Eventually(t, func() bool { return ran.Load() }, time.Second, 100*time.Millisecond)

	cancel()
	assert.ErrorIs(t, <-ch, context.Canceled)
}
