package scraper_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestWriter(t *testing.T) {
	p := poller{ch: make(chan solaredge.Update)}
	s := store{}
	w := scraper.Writer{
		Store: &s,
		TadoGetter: tadoClient{weatherInfo: tado.WeatherInfo{
			OutsideTemperature: tado.Temperature{Celsius: 23.0},
			SolarIntensity:     tado.Percentage{Percentage: 75},
			WeatherState:       tado.Value{Value: "SUNNY"},
		}},
		Poller:   &p,
		Interval: 100 * time.Millisecond,
		Logger:   slog.Default(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- w.Run(ctx) }()

	p.ch <- emptyUpdate
	assert.Never(t, s.hasData.Load, 500*time.Millisecond, time.Millisecond)

	p.ch <- testUpdate
	assert.Eventually(t, s.hasData.Load, time.Second, time.Millisecond)
	cancel()

	assert.NoError(t, <-ch)
	assert.Equal(t, "SUNNY", s.measurement.Weather)
	assert.Equal(t, 75.0, s.measurement.Intensity)
	assert.Equal(t, 3000.0, s.measurement.Power)
}

func TestWriter_TadoFailure(t *testing.T) {
	p := poller{ch: make(chan solaredge.Update)}
	s := store{}
	w := scraper.Writer{
		Store:      &s,
		TadoGetter: tadoClient{err: errors.New("failure")},
		Poller:     &p,
		Interval:   100 * time.Millisecond,
		Logger:     slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})),
	}

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- w.Run(ctx) }()

	p.ch <- testUpdate
	assert.Never(t, s.hasData.Load, 200*time.Millisecond, 50*time.Millisecond)
	cancel()

	assert.NoError(t, <-ch)
}
