package scraper

import (
	"context"
	solaredge2 "github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/poller/solaredge"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func VarP[T any](t T) *T {
	return &t
}

func TestWriter(t *testing.T) {
	s := store{}
	solarUpdate := fakePublisher[solaredge.Update]{ch: make(chan solaredge.Update)}
	tadoUpdate := fakePublisher[*tado.Weather]{ch: make(chan *tado.Weather)}

	w := Writer{
		Store:     &s,
		SolarEdge: solarUpdate,
		Tado:      tadoUpdate,
		Interval:  10 * time.Millisecond,
		Logger:    slog.Default(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() { errCh <- w.Run(ctx) }()

	solarUpdate.ch <- emptyUpdate
	assert.Never(t, s.hasData.Load, 100*time.Millisecond, time.Millisecond)

	solarUpdate.ch <- testUpdate
	assert.Never(t, s.hasData.Load, 100*time.Millisecond, time.Millisecond)

	tadoUpdate.ch <- &tado.Weather{
		SolarIntensity: &tado.PercentageDataPoint{Percentage: VarP(float32(75))},
		WeatherState:   &tado.WeatherStateDataPoint{Value: VarP(tado.SUN)},
	}
	assert.Eventually(t, s.hasData.Load, time.Second, time.Millisecond)

	cancel()
	assert.NoError(t, <-errCh)
	assert.Equal(t, "SUN", s.measurement.Weather)
	assert.Equal(t, 75.0, s.measurement.Intensity)
	assert.Equal(t, 1500.0, s.measurement.Power)
}

var _ Store = &store{}

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

type fakePublisher[T any] struct {
	ch chan T
}

func (f fakePublisher[T]) Subscribe() chan T {
	return f.ch
}

func (f fakePublisher[T]) Unsubscribe(_ chan T) {
	return
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
