package solaredge

import (
	"context"
	"github.com/clambin/solaredge"
	"time"
)

type Update []SiteUpdate

type SiteUpdate struct {
	ID              int
	Name            string
	PowerOverview   solaredge.PowerOverview
	InverterUpdates []InverterUpdate
}

type InverterUpdate struct {
	Name         string
	SerialNumber string
	Telemetry    solaredge.InverterTelemetry
}

type Client struct {
	SolarEdge solaredge.Client
}

func (c Client) GetUpdate(ctx context.Context) (Update, error) {
	sites, err := c.SolarEdge.GetSites(ctx)
	if err != nil {
		return Update{}, err
	}

	update := make(Update, len(sites))
	for s := range sites {
		siteUpdate := SiteUpdate{
			ID:   sites[s].ID,
			Name: sites[s].Name,
		}

		siteUpdate.PowerOverview, err = sites[s].GetPowerOverview(ctx)
		if err != nil {
			return Update{}, err
		}

		inverters, err := sites[s].GetInverters(ctx)
		if err != nil {
			return Update{}, err
		}

		end := time.Now()
		start := end.Add(-10 * time.Minute)

		inverterUpdates := make([]InverterUpdate, len(inverters))
		for i := range inverters {
			inverterUpdates[i].Name = inverters[i].Name
			inverterUpdates[i].SerialNumber = inverters[i].GetSerialNumber()

			telemetry, err := inverters[i].GetTelemetry(ctx, start, end)
			if err != nil {
				return Update{}, err
			}
			if len(telemetry) > 0 {
				inverterUpdates[i].Telemetry = telemetry[len(telemetry)-1]
			}
		}
		siteUpdate.InverterUpdates = inverterUpdates
		update[s] = siteUpdate
	}

	return update, nil
}
