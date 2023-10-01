package scraper

import (
	"context"
	"github.com/clambin/solaredge"
)

type PowerInfo struct {
	Power float64
}

type PowerGetter interface {
	GetPowerOverview(ctx context.Context) (solaredge.PowerOverview, error)
}

type PowerScraper struct {
	PowerGetter
	publisher[PowerInfo]
}

func (s *PowerScraper) Poll(ctx context.Context) error {
	var info PowerInfo
	overview, err := s.GetPowerOverview(ctx)
	if err == nil {
		info.Power = overview.CurrentPower.Power
		s.Publish(info)
	}
	return err
}
