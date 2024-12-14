package testutils

import (
	solaredge2 "github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/poller/solaredge"
)

type FakePublisher[T any] struct {
	Ch chan T
}

func (f FakePublisher[T]) Subscribe() chan T {
	return f.Ch
}

func (f FakePublisher[T]) Unsubscribe(_ chan T) {
}

var (
	TestUpdate = solaredge.Update{
		solaredge.SiteUpdate{
			ID:   1,
			Name: "foo",
			PowerOverview: solaredge2.PowerOverview{
				LastYearData:  solaredge2.EnergyOverview{Energy: 1000},
				LastMonthData: solaredge2.EnergyOverview{Energy: 100},
				LastDayData:   solaredge2.EnergyOverview{Energy: 10},
				CurrentPower:  solaredge2.CurrentPower{Power: 3000},
			},
			InverterUpdates: []solaredge.InverterUpdate{
				{
					Name:         "inv1",
					SerialNumber: "1234",
					Telemetry: solaredge2.InverterTelemetry{
						L1Data: struct {
							AcCurrent     float64 `json:"acCurrent"`
							AcFrequency   float64 `json:"acFrequency"`
							AcVoltage     float64 `json:"acVoltage"`
							ActivePower   float64 `json:"activePower"`
							ApparentPower float64 `json:"apparentPower"`
							CosPhi        float64 `json:"cosPhi"`
							ReactivePower float64 `json:"reactivePower"`
						}(struct {
							AcCurrent     float64
							AcFrequency   float64
							AcVoltage     float64
							ActivePower   float64
							ApparentPower float64
							CosPhi        float64
							ReactivePower float64
						}{
							AcCurrent: 10,
							AcVoltage: 240,
						}),
						DcVoltage:        400,
						PowerLimit:       1,
						Temperature:      40,
						TotalActivePower: 9999,
						TotalEnergy:      8888,
					},
				},
			},
		},
	}

	EmptyUpdate = solaredge.Update{
		solaredge.SiteUpdate{
			ID:   1,
			Name: "foo",
			PowerOverview: solaredge2.PowerOverview{
				LastYearData:  solaredge2.EnergyOverview{Energy: 1000},
				LastMonthData: solaredge2.EnergyOverview{Energy: 100},
				LastDayData:   solaredge2.EnergyOverview{Energy: 10},
				CurrentPower:  solaredge2.CurrentPower{Power: 0},
			},
		},
	}
)
