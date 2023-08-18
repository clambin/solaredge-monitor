package solaredgeScraper_test

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector/solaredgeScraper"
	"github.com/clambin/solaredge-monitor/collector/solaredgeScraper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFetcher_Run(t *testing.T) {
	response := solaredge.PowerOverview{
		LastUpdateTime: solaredge.Time{},
		LifeTimeData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 10000,
		},
		LastYearData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 1000,
		},
		LastMonthData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 100,
		},
		LastDayData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 10,
		},
		CurrentPower: struct {
			Power float64 `json:"power"`
		}{
			Power: 3400,
		},
	}

	site := mocks.NewSite(t)
	site.EXPECT().GetPowerOverview(mock.Anything).Return(response, nil)

	ch := make(chan solaredgeScraper.Info)
	f := solaredgeScraper.Fetcher{Site: site}
	go f.Run(context.Background(), time.Millisecond, ch)

	info := <-ch
	assert.Equal(t, solaredgeScraper.Info{Power: 3400}, info)
}
