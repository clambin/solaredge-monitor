package sampler

import (
	"context"
	"errors"
	"github.com/clambin/solaredge"
	"time"
)

type SolarEdgeSampler struct {
	solaredge.API
	SiteID int
}

var _ Sampler = &SolarEdgeSampler{}

func (s *SolarEdgeSampler) Sample(ctx context.Context) (sample Sample, err error) {
	if err = s.setSiteID(ctx); err != nil {
		return
	}

	var current float64
	if _, _, _, _, current, err = s.GetPowerOverview(ctx, s.SiteID); err == nil {
		sample = Sample{
			Timestamp: time.Now(),
			Value:     current,
		}
	}
	return
}

func (s *SolarEdgeSampler) setSiteID(ctx context.Context) (err error) {
	if s.SiteID != 0 {
		return
	}

	var siteIDs []int
	if siteIDs, err = s.GetSiteIDs(ctx); err != nil {
		return
	}
	if len(siteIDs) == 0 {
		return errors.New("solaredge: no sites appear to be registered")
	}

	s.SiteID = siteIDs[0]
	return
}
