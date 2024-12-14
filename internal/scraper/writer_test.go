package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/poller/solaredge"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/testutils"
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
	solarUpdate := testutils.FakePublisher[solaredge.Update]{Ch: make(chan solaredge.Update)}
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

	solarUpdate.Ch <- testutils.EmptyUpdate
	assert.Never(t, s.hasData.Load, 100*time.Millisecond, time.Millisecond)

	solarUpdate.Ch <- testutils.TestUpdate
	assert.Never(t, s.hasData.Load, 100*time.Millisecond, time.Millisecond)

	tadoUpdate.Ch <- &tado.Weather{
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
