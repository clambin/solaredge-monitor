package publisher

import (
	"context"
	"fmt"
	solaredge "github.com/clambin/solaredge/v2"
	"time"
)

type SolarEdgeUpdater struct {
	SolarEdgeClient
}

type SolarEdgeClient interface {
	GetSites(ctx context.Context) (solaredge.GetSitesResponse, error)
	GetPowerOverview(ctx context.Context, id int) (solaredge.GetPowerOverviewResponse, error)
	GetComponents(ctx context.Context, id int) (solaredge.GetComponentsResponse, error)
	GetInverterTechnicalData(ctx context.Context, id int, serialNr string, startTime time.Time, endTime time.Time) (solaredge.GetInverterTechnicalDataResponse, error)
}

func (c SolarEdgeUpdater) GetUpdate(ctx context.Context) (SolarEdgeUpdate, error) {
	sites, err := c.GetSites(ctx)
	if err != nil {
		return SolarEdgeUpdate{}, err
	}

	update := make(SolarEdgeUpdate, len(sites.Sites.Site))
	for i, site := range sites.Sites.Site {
		siteUpdate := SiteUpdate{
			ID:   site.Id,
			Name: site.Name,
		}
		if siteUpdate.PowerOverview, siteUpdate.InverterUpdates, err = c.getSiteUpdate(ctx, site.Id); err != nil {
			return SolarEdgeUpdate{}, err
		}
		update[i] = siteUpdate
	}

	return update, nil
}

func (c SolarEdgeUpdater) getSiteUpdate(ctx context.Context, id int) (solaredge.PowerOverview, []InverterUpdate, error) {
	powerOverview, err := c.GetPowerOverview(ctx, id)
	if err != nil {
		return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get power overview: %w", err)
	}

	inverters, err := c.GetComponents(ctx, id)
	if err != nil {
		return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get inverters: %w", err)
	}

	endTime := time.Now()
	startTime := endTime.Add(-10 * time.Minute)

	inverterUpdates := make([]InverterUpdate, len(inverters.Reporters.List))
	for i, inverter := range inverters.Reporters.List {
		inverterUpdates[i] = InverterUpdate{
			Name:         inverter.Name,
			SerialNumber: inverter.SerialNumber,
		}

		telemetry, err := c.GetInverterTechnicalData(ctx, id, inverter.SerialNumber, startTime, endTime)
		if err != nil {
			return solaredge.PowerOverview{}, nil, fmt.Errorf("unable to get telemetry for inverter %q: %w", inverter.Name, err)
		}
		if n := len(telemetry.Data.Telemetries); n > 0 {
			inverterUpdates[i].Telemetry = telemetry.Data.Telemetries[n-1]
		}
	}
	return powerOverview.Overview, inverterUpdates, nil
}
