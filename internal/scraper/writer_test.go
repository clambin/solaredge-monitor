package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func VarP[T any](t T) *T {
	return &t
}

func TestWriter(t *testing.T) {
	s := store{}
	solarUpdate := testutils.FakePublisher[publisher.SolarEdgeUpdate]{Ch: make(chan publisher.SolarEdgeUpdate)}
	tadoUpdate := testutils.FakePublisher[*tado.Weather]{Ch: make(chan *tado.Weather)}

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

	solarUpdate.Ch <- testutils.TestUpdate
	tadoUpdate.Ch <- &tado.Weather{
		SolarIntensity: &tado.PercentageDataPoint{Percentage: VarP(float32(75))},
		WeatherState:   &tado.WeatherStateDataPoint{Value: VarP(tado.SUN)},
	}

	assert.Eventually(t, s.hasData.Load, time.Second, time.Millisecond)

	cancel()
	assert.NoError(t, <-errCh)
	assert.Equal(t, "SUN", s.measurement.Weather)
	assert.Equal(t, 75.0, s.measurement.Intensity)
	assert.Equal(t, 3000.0, s.measurement.Power)
}

func TestWriter_store(t *testing.T) {
	tests := []struct {
		name    string
		solar   []publisher.SolarEdgeUpdate
		tado    []*tado.Weather
		hasData assert.BoolAssertionFunc
	}{
		{
			name:  "no power: no update",
			solar: []publisher.SolarEdgeUpdate{testutils.EmptyUpdate},
			tado: []*tado.Weather{{
				SolarIntensity: &tado.PercentageDataPoint{Percentage: VarP(float32(75))},
				WeatherState:   &tado.WeatherStateDataPoint{Value: VarP(tado.SUN)},
			}},
			hasData: assert.False,
		},
		{
			name:  "no solaredge update: no update",
			solar: []publisher.SolarEdgeUpdate{},
			tado: []*tado.Weather{{
				SolarIntensity: &tado.PercentageDataPoint{Percentage: VarP(float32(75))},
				WeatherState:   &tado.WeatherStateDataPoint{Value: VarP(tado.SUN)},
			}},
			hasData: assert.False,
		},
		{
			name:    "no tado update: no update",
			solar:   []publisher.SolarEdgeUpdate{testutils.TestUpdate},
			tado:    []*tado.Weather{},
			hasData: assert.False,
		},
		{
			name:  "power: update",
			solar: []publisher.SolarEdgeUpdate{testutils.TestUpdate},
			tado: []*tado.Weather{{
				SolarIntensity: &tado.PercentageDataPoint{Percentage: VarP(float32(75))},
				WeatherState:   &tado.WeatherStateDataPoint{Value: VarP(tado.SUN)},
			}},
			hasData: assert.True,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := store{}
			w := Writer{
				Store:  &s,
				Logger: discardLogger,
			}
			for _, u := range tt.solar {
				w.processSolarEdgeUpdate(u)
			}
			for _, u := range tt.tado {
				w.processTadoUpdate(u)
			}
			assert.NoError(t, w.store())
			tt.hasData(t, s.hasData.Load())
			if s.hasData.Load() {
				assert.Zero(t, w.solarIntensity.len())
				assert.Zero(t, w.power.len())
				assert.Empty(t, 0, w.weatherStates)
			}
		})
	}
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
