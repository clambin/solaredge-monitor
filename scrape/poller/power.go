package poller

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/scrape/collector"
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

func (poller *SolarEdgePoller) poll(ctx context.Context) (result float64, err error) {
	var sites []int
	sites, err = poller.API.GetSiteIDs(ctx)

	if err == nil {
		for _, siteID := range sites {
			var current float64
			_, _, _, _, current, err = poller.API.GetPowerOverview(ctx, siteID)

			result += current
		}
	}

	log.WithError(err).WithField("power", result).Debug("power polled")
	return
}
