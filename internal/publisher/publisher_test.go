package publisher

import (
	"context"
	"github.com/clambin/go-common/pubsub"
	"github.com/clambin/solaredge-monitor/internal/publisher/solaredge"
	"github.com/clambin/tado/v2"
	"io"
	"log/slog"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestPublisher_SolarEdge(t *testing.T) {
	p := Publisher[solaredge.Update]{
		Updater:   fakeSolarEdgeClient{},
		Interval:  100 * time.Millisecond,
		Logger:    discardLogger,
		Publisher: pubsub.Publisher[solaredge.Update]{},
	}
	ch := p.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error)
	go func() { errCh <- p.Run(ctx) }()

	<-ch
	<-ch
}

var _ Updater[solaredge.Update] = fakeSolarEdgeClient{}

type fakeSolarEdgeClient struct{}

func (f fakeSolarEdgeClient) GetUpdate(_ context.Context) (solaredge.Update, error) {
	return solaredge.Update{}, nil
}

func TestPublisher_Tado(t *testing.T) {
	p := Publisher[*tado.Weather]{
		Updater:   fakeTadoClient{},
		Interval:  100 * time.Millisecond,
		Logger:    discardLogger,
		Publisher: pubsub.Publisher[*tado.Weather]{},
	}
	ch := p.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error)
	go func() { errCh <- p.Run(ctx) }()

	<-ch
	<-ch
}

var _ Updater[*tado.Weather] = fakeTadoClient{}

type fakeTadoClient struct{}

func (f fakeTadoClient) GetUpdate(_ context.Context) (*tado.Weather, error) {
	return &tado.Weather{
		OutsideTemperature: &tado.TemperatureDataPoint{Celsius: varP(float32(18))},
		SolarIntensity:     &tado.PercentageDataPoint{Percentage: varP(float32(75))},
		WeatherState:       &tado.WeatherStateDataPoint{Value: varP(tado.SUN)},
	}, nil
}

func varP[T any](t T) *T { return &t }
