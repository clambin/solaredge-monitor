package poller

import (
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/tado"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type TadoPoller struct {
	BasePoller
	tado.API
}

func NewTadoPoller(username, password string, summary chan collector.Metric, pollInterval time.Duration) *TadoPoller {
	c := &TadoPoller{
		API: &tado.APIClient{
			Username:   username,
			Password:   password,
			HTTPClient: &http.Client{},
		},
	}
	c.BasePoller = *NewBasePoller(pollInterval, c.poll, summary)
	return c
}

func (poller *TadoPoller) poll() (result float64, err error) {
	var weatherInfo tado.WeatherInfo
	weatherInfo, err = poller.API.GetWeatherInfo()

	if err == nil {
		result = weatherInfo.SolarIntensity.Percentage
	}
	log.WithError(err).WithField("intensity", result).Debug("solar intensity polled")
	return
}
