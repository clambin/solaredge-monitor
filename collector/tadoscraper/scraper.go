package tadoscraper

import (
	"context"
	tado2 "github.com/clambin/solaredge-monitor/tado"
	"time"
)

type Info struct {
	SolarIntensity float64
	Temperature    float64
	Weather        string
}

type Fetcher struct {
	tado2.API
}

func (f *Fetcher) Run(ctx context.Context, interval time.Duration, ch chan<- Info) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if info, err := f.fetch(ctx); err == nil {
				ch <- info
			}
		}
	}
}

func (f *Fetcher) fetch(ctx context.Context) (Info, error) {
	var info Info
	weatherInfo, err := f.API.GetWeatherInfo(ctx)
	if err == nil {
		info.SolarIntensity = weatherInfo.SolarIntensity.Percentage
		info.Temperature = weatherInfo.OutsideTemperature.Celsius
		info.Weather = weatherInfo.WeatherState.Value
	}
	return info, err
}
