package solaredge

import (
	"context"
	"github.com/clambin/solaredge"
)

// API interface abstracts the solaredge API, so we can mock it during unit testing
//
//go:generate mockery --name API
type API interface {
	GetPowerOverview(context.Context) (solaredge.PowerOverview, error)
}
