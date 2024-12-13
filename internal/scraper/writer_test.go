package scraper_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func VarP[T any](t T) *T {
	return &t
}

func TestWriter(t *testing.T) {
	p := poller{ch: make(chan solaredge.Update)}
	s := store{}
	w := scraper.Writer{
		Store: &s,
		TadoGetter: tadoClient{weatherInfo: tado.Weather{
			OutsideTemperature: &tado.TemperatureDataPoint{Celsius: VarP[float32](23.0)},
			SolarIntensity:     &tado.PercentageDataPoint{Percentage: VarP[float32](75)},
			WeatherState:       &tado.WeatherStateDataPoint{Value: VarP[tado.WeatherState](tado.SUN)},
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
	assert.Equal(t, "SUN", s.measurement.Weather)
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
