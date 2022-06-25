package sampler_test

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/scrape/sampler"
	"github.com/clambin/tado"
	"github.com/clambin/tado/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTadoClient(t *testing.T) {
	api := &mocks.API{}
	c := sampler.TadoSampler{
		API: api,
	}

	api.On("GetWeatherInfo", mock.AnythingOfType("*context.emptyCtx")).Return(tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 15.0},
		SolarIntensity:     tado.Percentage{Percentage: 75.0},
		WeatherState:       tado.Value{Value: "SUNNY"},
	}, nil).Once()

	sample, err := c.Sample(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 75.0, sample.Value)

	api.On("GetWeatherInfo", mock.AnythingOfType("*context.emptyCtx")).Return(tado.WeatherInfo{}, errors.New("fail")).Once()
	_, err = c.Sample(context.Background())
	assert.Error(t, err)

	mock.AssertExpectationsForObjects(t, api)
}
