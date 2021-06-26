package poller_test

import (
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/poller"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSolarEdgePoller(t *testing.T) {
	summary := make(chan collector.Metric)
	p := poller.NewSolarEdgePoller("", summary, 50*time.Millisecond)
	p.API = &SolarEdgeMock{}
	go p.Run()

	received := <-summary

	assert.Equal(t, 1.0, received.Value)

	p.Stop <- struct{}{}
}

type SolarEdgeMock struct{}

func (api *SolarEdgeMock) GetSiteIDs() (siteIDs []int, err error) {
	return []int{1}, nil
}

func (api *SolarEdgeMock) GetPower(_ int, startTime, endTime time.Time) (entries []solaredge.PowerMeasurement, err error) {
	var value float64

	for startTime.Before(endTime) {
		entries = append(entries, solaredge.PowerMeasurement{
			Time:  startTime,
			Value: value,
		})
		startTime = startTime.Add(15 * time.Minute)
		value += 100.0
	}
	return
}

func (api *SolarEdgeMock) GetPowerOverview(_ int) (lifeTime, lastYear, lastMonth, lastDay, current float64, err error) {
	return 10000.0, 1000.0, 100.0, 10.0, 1.0, nil
}
