package scraper

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/scraper/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"testing"
	"time"
)

func TestPowerScraper(t *testing.T) {
	g := newMockPowerGetter(t)
	s := PowerScraper{PowerGetter: g}
	d := daemon{
		Logger:   slog.Default(),
		Interval: time.Millisecond,
		Poller:   &s,
	}
	ch := s.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- d.Run(ctx)
	}()

	assert.Equal(t, PowerInfo{Power: 1000}, <-ch)
	cancel()
	assert.ErrorIs(t, <-errCh, context.Canceled)
}

func newMockPowerGetter(t *testing.T) PowerGetter {
	overview := solaredge.PowerOverview{CurrentPower: struct {
		Power float64 `json:"power"`
	}{Power: 1000}}
	g := mocks.NewPowerGetter(t)
	g.EXPECT().GetPowerOverview(mock.Anything).Return(overview, nil)
	return g
}
