package solaredgescraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/solaredge"
	"time"
)

type Info struct {
	Power float64
}

type Fetcher struct {
	solaredge.API
	siteID int
}

func (f *Fetcher) Run(ctx context.Context, interval time.Duration, ch chan<- Info) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if info, err := f.fetch(ctx); err == nil {
				ch <- info
			}
		}
	}
}

func (f *Fetcher) fetch(ctx context.Context) (Info, error) {
	var info Info

	overview, err := f.API.GetPowerOverview(ctx)
	if err == nil {
		info.Power = overview.CurrentPower.Power
	}
	return info, err
}
