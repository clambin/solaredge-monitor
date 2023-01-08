package solaredgescraper

import (
	"context"
	"github.com/clambin/solaredge"
	"golang.org/x/exp/slog"
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
				ch <- *info
			}
		}
	}
}

func (f *Fetcher) fetch(ctx context.Context) (*Info, error) {
	if f.siteID == 0 {
		var err error
		f.siteID, err = f.getSiteID(ctx)
		if err != nil {
			return nil, err
		}
	}
	var info *Info
	_, _, _, _, current, err := f.API.GetPowerOverview(ctx, f.siteID)
	if err == nil {
		info = &Info{
			Power: current,
		}
	}
	return info, err
}

func (f *Fetcher) getSiteID(ctx context.Context) (int, error) {
	var siteID int
	siteIDs, err := f.API.GetSiteIDs(ctx)
	if err == nil {
		if len(siteIDs) > 1 {
			slog.Warn("found multiple siteIDs. picking first one")
		}
		siteID = siteIDs[0]
	}

	return siteID, err
}
