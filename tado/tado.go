package tado

import (
	"context"
	"github.com/clambin/tado"
)

// API interface abstracts the tado API, so we can mock it during unit testing
//
//go:generate mockery --name API
type API interface {
	GetWeatherInfo(ctx context.Context) (tado.WeatherInfo, error)
}
