package publisher

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge"
	solaredge2 "github.com/clambin/solaredge-monitor/internal/publisher/solaredge"
	"time"
)

type SolarEdgeUpdater struct {
	SolarEdge solaredge.Client
}

func (c SolarEdgeUpdater) GetUpdate(ctx context.Context) (solaredge2.Update, error) {
	sites, err := c.SolarEdge.GetSites(ctx)
	if err != nil {
		return solaredge2.Update{}, err
	}

	update := make(solaredge2.Update, len(sites))
	for i, site := range sites {
		siteUpdate := solaredge2.SiteUpdate{
			ID:   site.ID,
			Name: site.Name,
		}
		if siteUpdate.PowerOverview, siteUpdate.InverterUpdates, err = getSiteUpdate(ctx, site); err != nil {
			return solaredge2.Update{}, err
		}
		update[i] = siteUpdate
	}

	return update, nil
}

func getSiteUpdate(ctx context.Context, site solaredge.Site) (solaredge.PowerOverview, []solaredge2.InverterUpdate, error) {
	powerOverview, err := site.GetPowerOverview(ctx)
	if err != nil {
		return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get power overview: %w", err)
	}

	inverters, err := site.GetInverters(ctx)
	if err != nil {
		return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get inverters: %w", err)
	}

	end := time.Now()
	start := end.Add(-10 * time.Minute)

	inverterUpdates := make([]solaredge2.InverterUpdate, len(inverters))
	for i := range inverters {
		inverterUpdates[i].Name = inverters[i].Name
		inverterUpdates[i].SerialNumber = inverters[i].GetSerialNumber()
		telemetry, err := inverters[i].GetTelemetry(ctx, start, end)
		if err != nil {
			return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get telemetry for inverter %q: %w", inverters[i].Name, err)
		}
		if len(telemetry) > 0 {
			inverterUpdates[i].Telemetry = telemetry[len(telemetry)-1]
		}
	}
	return powerOverview, inverterUpdates, nil
}
