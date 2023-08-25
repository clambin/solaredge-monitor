package scraper

import (
	"context"
	"github.com/clambin/tado"
)

type WeatherInfo struct {
	SolarIntensity float64
	Temperature    float64
	Weather        string
}

type WeatherGetter interface {
	GetWeatherInfo(ctx context.Context) (tado.WeatherInfo, error)
}

type WeatherScraper struct {
	WeatherGetter
	publisher[WeatherInfo]
}

func (s *WeatherScraper) Poll(ctx context.Context) error {
	var info WeatherInfo
	weatherInfo, err := s.GetWeatherInfo(ctx)
	if err == nil {
		info.SolarIntensity = weatherInfo.SolarIntensity.Percentage
		info.Temperature = weatherInfo.OutsideTemperature.Celsius
		info.Weather = weatherInfo.WeatherState.Value
		s.Publish(info)
	}
	return err
}
