package solaredgescraper_test

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector/solaredgescraper"
	"github.com/clambin/solaredge-monitor/solaredge/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFetcher_Run(t *testing.T) {
	api := mocks.NewAPI(t)
	api.On("GetPowerOverview", mock.AnythingOfType("*context.emptyCtx")).Return(solaredge.PowerOverview{
		CurrentPower: struct {
			Power float64 `json:"power"`
		}{
			Power: 3200,
		},
	}, nil)

	ch := make(chan solaredgescraper.Info)
	f := solaredgescraper.Fetcher{API: api}
	go f.Run(context.Background(), time.Millisecond, ch)

	info := <-ch
	assert.Equal(t, solaredgescraper.Info{Power: 3200}, info)
}
