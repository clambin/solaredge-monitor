package sampler

import (
	"context"
	"github.com/clambin/tado"
	"time"
)

type TadoSampler struct {
	tado.API
}

var _ Sampler = &TadoSampler{}

func (t *TadoSampler) Sample(ctx context.Context) (sample Sample, err error) {
	var weatherInfo tado.WeatherInfo
	if weatherInfo, err = t.GetWeatherInfo(ctx); err == nil {
		sample = Sample{
			Timestamp: time.Now(),
			Value:     weatherInfo.SolarIntensity.Percentage,
		}
	}
	return
}
