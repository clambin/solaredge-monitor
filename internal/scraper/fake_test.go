package scraper_test

import (
	"context"
	solaredge2 "github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/tado/v2"
	"sync"
	"sync/atomic"
)

var _ scraper.Publisher[solaredge.Update] = poller{}

type poller struct {
	ch chan solaredge.Update
}

func (p poller) Subscribe() chan solaredge.Update {
	return p.ch
}

func (p poller) Unsubscribe(ch chan solaredge.Update) {
	if ch != p.ch {
		panic("unexpected channel")
	}
}

var _ scraper.SolarEdgeGetter = client{}

type client struct {
	update solaredge.Update
}

func (c client) GetUpdate(_ context.Context) (solaredge.Update, error) {
	return c.update, nil
}

var _ scraper.Store = &store{}

type store struct {
	hasData     atomic.Bool
	lock        sync.Mutex
	measurement repository.Measurement
}

func (s *store) Store(measurement repository.Measurement) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.measurement = measurement
	s.hasData.Store(true)
	return nil
}

var _ scraper.TadoGetter = tadoClient{}

type tadoClient struct {
	weatherInfo tado.Weather
	err         error
}

func (t tadoClient) GetWeatherWithResponse(_ context.Context, _ tado.HomeId, _ ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error) {
	resp := tado.GetWeatherResponse{
		JSON200: &t.weatherInfo,
	}
	return &resp, t.err
}

var (
	testUpdate = solaredge.Update{
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

	emptyUpdate = solaredge.Update{
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
