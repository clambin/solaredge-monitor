package testutils

import (
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge/v2"
)

type FakePublisher[T any] struct {
	Ch chan T
}

func (f FakePublisher[T]) Subscribe() <-chan T {
	return f.Ch
}

func (f FakePublisher[T]) Unsubscribe(_ <-chan T) {
}

var (
	TestUpdate = publisher.SolarEdgeUpdate{
		{
			ID:   1,
			Name: "foo",
			PowerOverview: solaredge.PowerOverview{
				LastYearData:  solaredge.EnergyOverview{Energy: 1000},
				LastMonthData: solaredge.EnergyOverview{Energy: 100},
				LastDayData:   solaredge.EnergyOverview{Energy: 10},
				CurrentPower:  solaredge.CurrentPower{Power: 3000},
			},
			InverterUpdates: []publisher.InverterUpdate{
				{
					Name:         "inv1",
					SerialNumber: "1234",
					Telemetry: solaredge.InverterTelemetry{
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

	EmptyUpdate = publisher.SolarEdgeUpdate{
		{
			ID:   1,
			Name: "foo",
			PowerOverview: solaredge.PowerOverview{
				LastYearData:  solaredge.EnergyOverview{Energy: 1000},
				LastMonthData: solaredge.EnergyOverview{Energy: 100},
				LastDayData:   solaredge.EnergyOverview{Energy: 10},
				CurrentPower:  solaredge.CurrentPower{Power: 0},
			},
		},
	}
)
