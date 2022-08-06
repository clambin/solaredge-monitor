package scraper

import (
	"context"
	"github.com/clambin/tado"
	"time"
)

type TadoScraper struct {
	tado.API
}

var _ Scraper = &TadoScraper{}

func (t *TadoScraper) Scrape(ctx context.Context) (sample Sample, err error) {
	var weatherInfo tado.WeatherInfo
	if weatherInfo, err = t.GetWeatherInfo(ctx); err == nil {
		sample = Sample{
			Timestamp: time.Now(),
			Value:     weatherInfo.SolarIntensity.Percentage,
		}
	}
	return
}
