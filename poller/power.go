package poller

import (
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type SolarEdgePoller struct {
	BasePoller
	solaredge.API
}

func NewSolarEdgePoller(token string, summary chan collector.Metric, pollInterval time.Duration) *SolarEdgePoller {
	c := &SolarEdgePoller{
		API: solaredge.NewClient(token, &http.Client{}),
	}
	c.BasePoller = *NewBasePoller(pollInterval, c.poll, summary)
	return c
}

func (poller *SolarEdgePoller) poll() (result float64, err error) {
	var sites []int
	sites, err = poller.API.GetSiteIDs()

	if err == nil {
		for _, siteID := range sites {
			var current float64
			_, _, _, _, current, err = poller.API.GetPowerOverview(siteID)

			result += current
		}
	}

	log.WithField("power", result).Debug("power polled")
	return
}
