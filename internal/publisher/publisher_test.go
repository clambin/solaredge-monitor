package publisher

import (
	"codeberg.org/clambin/go-common/pubsub"
	"context"
	"github.com/clambin/solaredge/v2"
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

	go func() { assert.NoError(t, p.Run(t.Context())) }()

	assert.Equal(t, SolarEdgeUpdate{{
		ID:   1,
		Name: "my home",
		PowerOverview: solaredge.PowerOverview{
			LifeTimeData:  solaredge.EnergyOverview{Energy: 1000},
			LastYearData:  solaredge.EnergyOverview{Energy: 100},
			LastMonthData: solaredge.EnergyOverview{Energy: 10},
			LastDayData:   solaredge.EnergyOverview{Energy: 1},
			CurrentPower:  solaredge.CurrentPower{Power: 100},
		},
		InverterUpdates: []InverterUpdate{{
			Name:         "foo",
			SerialNumber: "1234",
			Telemetry: solaredge.InverterTelemetry{
				L1Data:    solaredge.InverterTelemetryL1Data{AcCurrent: 1, AcVoltage: 220},
				DcVoltage: 380,
			},
		}},
	}}, <-ch)

	<-ch
}

func TestPublisher_Tado(t *testing.T) {
	p := Publisher[*tado.Weather]{
		Updater:   fakeTadoClient{},
		Interval:  100 * time.Millisecond,
		Logger:    discardLogger,
		Publisher: pubsub.Publisher[*tado.Weather]{},
	}
	ch := p.Subscribe()

	go func() { assert.NoError(t, p.Run(t.Context())) }()

	update := <-ch
	assert.Equal(t, float32(75), *update.SolarIntensity.Percentage)
	assert.Equal(t, float32(18), *update.OutsideTemperature.Celsius)
	assert.Equal(t, tado.SUN, *update.WeatherState.Value)
	<-ch
}

func TestPublisher_IsHealthy(t *testing.T) {
	p := Publisher[*tado.Weather]{Interval: 10 * time.Millisecond}
	assert.Error(t, p.IsHealthy(context.TODO()))
	p.lastUpdate.Store(time.Now())
	assert.NoError(t, p.IsHealthy(context.TODO()))
	assert.Eventually(t, func() bool { return p.IsHealthy(context.TODO()) != nil }, time.Second, p.Interval)
}

func TestPublisher_getSource(t *testing.T) {
	var p Publisher[*tado.Weather]
	assert.Equal(t, "Tado", p.getSource())
	var q Publisher[SolarEdgeUpdate]
	assert.Equal(t, "SolarEdge", q.getSource())
	var r Publisher[any]
	assert.Equal(t, "unknown source", r.getSource())
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
