package tadoscraper_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/collector/tadoscraper"
	"github.com/clambin/solaredge-monitor/collector/tadoscraper/mocks"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFetcher_Run(t *testing.T) {
	response := tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 18.5},
		SolarIntensity:     tado.Percentage{Percentage: 75.0},
		WeatherState:       tado.Value{Value: "SUNNY"},
	}

	api := mocks.NewAPI(t)
	api.EXPECT().GetWeatherInfo(mock.Anything).Return(response, nil)

	ch := make(chan tadoscraper.Info)
	f := tadoscraper.Fetcher{API: api}
	go f.Run(context.Background(), time.Millisecond, ch)

	info := <-ch
	assert.Equal(t, tadoscraper.Info{SolarIntensity: 75, Temperature: 18.5, Weather: "SUNNY"}, info)
}
