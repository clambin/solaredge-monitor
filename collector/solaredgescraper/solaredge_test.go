package solaredgescraper_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/collector/solaredgescraper"
	"github.com/clambin/solaredge/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestFetcher_Run(t *testing.T) {
	api := mocks.NewAPI(t)
	api.On("GetSiteIDs", mock.AnythingOfType("*context.emptyCtx")).Return([]int{100, 102}, nil)
	api.On("GetPowerOverview", mock.AnythingOfType("*context.emptyCtx"), 100).Return(0.0, 0.0, 0.0, 0.0, 3200.0, nil)

	ch := make(chan solaredgescraper.Info)
	f := solaredgescraper.Fetcher{API: api}
	go f.Run(context.Background(), time.Millisecond, ch)

	info := <-ch
	assert.Equal(t, solaredgescraper.Info{Power: 3200}, info)
}
