package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/scraper/mocks"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"testing"
	"time"
)

func TestWeatherScraper(t *testing.T) {
	g := newMockWeatherGetter(t)
	s := WeatherScraper{WeatherGetter: g}
	d := daemon{
		Logger:   slog.Default(),
		Interval: time.Millisecond,
		Poller:   &s,
	}
	ch := s.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- d.Run(ctx)
	}()

	assert.Equal(t, WeatherInfo{
		SolarIntensity: 75,
		Temperature:    25,
		Weather:        "SUNNY",
	}, <-ch)
	cancel()
	assert.ErrorIs(t, <-errCh, context.Canceled)
}

func newMockWeatherGetter(t *testing.T) WeatherGetter {
	overview := tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 25},
		SolarIntensity:     tado.Percentage{Percentage: 75},
		WeatherState:       tado.Value{Value: "SUNNY"},
	}
	g := mocks.NewWeatherGetter(t)
	g.EXPECT().GetWeatherInfo(mock.Anything).Return(overview, nil)
	return g
}
