package publisher

import (
	"context"
	"github.com/clambin/go-common/pubsub"
	v2 "github.com/clambin/solaredge/v2"
	"github.com/clambin/tado/v2"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestPublisher_SolarEdge(t *testing.T) {
	p := Publisher[SolarEdgeUpdate]{
		Updater:   SolarEdgeUpdater{SolarEdgeClient: fakeSolarEdgeClient{}},
		Interval:  100 * time.Millisecond,
		Logger:    discardLogger,
		Publisher: pubsub.Publisher[SolarEdgeUpdate]{},
	}
	ch := p.Subscribe()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() { errCh <- p.Run(ctx) }()

	assert.Equal(t, SolarEdgeUpdate{{
		ID:   1,
		Name: "my home",
		PowerOverview: v2.PowerOverview{
			LifeTimeData:  v2.EnergyOverview{Energy: 1000},
			LastYearData:  v2.EnergyOverview{Energy: 100},
			LastMonthData: v2.EnergyOverview{Energy: 10},
			LastDayData:   v2.EnergyOverview{Energy: 1},
			CurrentPower:  v2.CurrentPower{Power: 100},
		},
		InverterUpdates: []InverterUpdate{{
			Name:         "foo",
			SerialNumber: "1234",
			Telemetry: v2.InverterTelemetry{
				L1Data:    v2.InverterTelemetryL1Data{AcCurrent: 1, AcVoltage: 220},
				DcVoltage: 380,
			},
		}},
	}}, <-ch)

	<-ch

	cancel()
	assert.NoError(t, <-errCh)
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
	errCh := make(chan error)
	go func() { errCh <- p.Run(ctx) }()

	update := <-ch
	assert.Equal(t, float32(75), *update.SolarIntensity.Percentage)
	assert.Equal(t, float32(18), *update.OutsideTemperature.Celsius)
	assert.Equal(t, tado.SUN, *update.WeatherState.Value)
	<-ch

	cancel()
	assert.NoError(t, <-errCh)
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
