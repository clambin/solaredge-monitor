package publisher

import (
	solaredge "github.com/clambin/solaredge/v2"
)

type SolarEdgeUpdate []SiteUpdate

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
