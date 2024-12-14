package solaredge

import (
	"github.com/clambin/solaredge"
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
