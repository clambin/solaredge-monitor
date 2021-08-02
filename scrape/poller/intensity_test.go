package poller_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/poller"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTadoCollector(t *testing.T) {
	collect := make(chan collector.Metric)
	p := poller.NewTadoPoller("foo", "bar", collect, 50*time.Millisecond)
	p.API = &TadoMock{}
	ctx, cancel := context.WithCancel(context.Background())
	go p.Run(ctx)

	received := <-collect

	assert.Equal(t, 13.3, received.Value)

	cancel()
}

type TadoMock struct {
}

func (api *TadoMock) GetZones(_ context.Context) ([]tado.Zone, error) {
	return []tado.Zone{}, nil
}

func (api *TadoMock) GetZoneInfo(_ context.Context, _ int) (tado.ZoneInfo, error) {
	return tado.ZoneInfo{}, nil
}

func (api *TadoMock) GetWeatherInfo(_ context.Context) (tado.WeatherInfo, error) {
	return tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 3.4},
		SolarIntensity:     tado.Percentage{Percentage: 13.3},
		WeatherState:       tado.Value{Value: "CLOUDY_MOSTLY"},
	}, nil
}

func (api *TadoMock) GetMobileDevices(_ context.Context) ([]tado.MobileDevice, error) {
	return []tado.MobileDevice{}, nil
}

func (api *TadoMock) SetZoneOverlay(_ context.Context, _ int, _ float64) error {
	return nil
}

func (api *TadoMock) SetZoneOverlayWithDuration(_ context.Context, _ int, _ float64, _ time.Duration) error {
	return nil
}

func (api *TadoMock) DeleteZoneOverlay(_ context.Context, _ int) error {
	return nil
}
