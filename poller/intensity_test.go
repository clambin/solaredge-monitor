package poller_test

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/poller"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTadoCollector(t *testing.T) {
	collect := make(chan collector.Metric)
	p := poller.NewTadoPoller("foo", "bar", collect, 50 * time.Millisecond)
	p.API = &TadoMock{}
	go p.Run()

	received := <-collect

	assert.Equal(t, 13.3, received.Value)

	p.Stop <- struct{}{}
}


type TadoMock struct {
}

func (api *TadoMock) GetZones() ([]tado.Zone, error) {
	return []tado.Zone{}, nil
}

func (api *TadoMock) GetZoneInfo(_ int) (tado.ZoneInfo, error) {
	return tado.ZoneInfo{}, nil
}

func (api *TadoMock) GetWeatherInfo() (tado.WeatherInfo, error) {
	return tado.WeatherInfo{
		OutsideTemperature: tado.Temperature{Celsius: 3.4},
		SolarIntensity:     tado.Percentage{Percentage: 13.3},
		WeatherState:       tado.Value{Value: "CLOUDY_MOSTLY"},
	}, nil
}

func (api *TadoMock) GetMobileDevices() ([]tado.MobileDevice, error) {
	return []tado.MobileDevice{}, nil
}

func (api *TadoMock) SetZoneOverlay(_ int, _ float64) error {
	return nil
}

func (api *TadoMock) DeleteZoneOverlay(_ int) error {
	return nil
}
