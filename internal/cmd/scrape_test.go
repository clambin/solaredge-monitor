package cmd

import (
	"context"
	"errors"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/clambin/solaredge/v2"
	"github.com/clambin/tado/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"net/http"
	"testing"
	"time"
)

func Test_runScrape(t *testing.T) {
	ctx := t.Context()
	store, connString, err := testutils.NewTestPostgresDB(ctx, "solaredge", "username", "password")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(store))
	})
	v := getViperFromViper(viper.GetViper())
	v.Set("database.url", connString)
	v.Set("polling.interval", time.Second)
	v.Set("scrape.interval", 2*time.Second)
	solarEdgeUpdater := fakeUpdater{SolarEdgeUpdate: publisher.SolarEdgeUpdate{{
		ID:   1,
		Name: "my home",
		PowerOverview: solaredge.PowerOverview{
			LastUpdateTime: solaredge.Time(time.Date(2024, time.December, 12, 12, 0, 0, 0, time.UTC)),
			LifeTimeData:   solaredge.EnergyOverview{Energy: 1000},
			LastYearData:   solaredge.EnergyOverview{Energy: 100},
			LastMonthData:  solaredge.EnergyOverview{Energy: 10},
			LastDayData:    solaredge.EnergyOverview{Energy: 1},
			CurrentPower:   solaredge.CurrentPower{Power: 500},
		},
	}}}
	tadoUpdater := publisher.TadoUpdater{Client: fakeTadoGetter{}}
	r := prometheus.NewPedanticRegistry()

	go func() {
		assert.NoError(t, runScrape(ctx, "dev", v, r, &solarEdgeUpdater, &tadoUpdater, discardLogger))
	}()

	dbc, err := repository.NewPostgresDB(connString)
	require.NoError(t, err)
	assert.Eventually(t, func() bool {
		rows, err := dbc.Get(time.Time{}, time.Time{})
		return err == nil && len(rows) > 0

	}, 10*time.Second, time.Millisecond)
}

func Test_getHomeId(t *testing.T) {
	type args struct {
		resp *tado.GetMeResponse
		err  error
	}
	type want struct {
		homeId tado.HomeId
		err    assert.ErrorAssertionFunc
	}
	tests := []struct {
		name string
		args
		want
	}{
		{
			name: "success",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{{Id: pointer(tado.HomeId(1))}}},
				},
				err: nil,
			},
			want: want{homeId: tado.HomeId(1), err: assert.NoError},
		},
		{
			name: "error",
			args: args{
				err: errors.New("some error"),
			},
			want: want{err: assert.Error},
		},
		{
			name: "no homes",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{}},
				},
				err: nil,
			},
			want: want{err: assert.Error},
		},
		{
			name: "more than one home",
			args: args{
				resp: &tado.GetMeResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					JSON200:      &tado.User{Homes: &[]tado.HomeBase{{Id: pointer(tado.HomeId(1))}, {Id: pointer(tado.HomeId(2))}}},
				},
				err: nil,
			},
			want: want{homeId: 1, err: assert.NoError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fakeMeGetter{tt.args.resp, tt.args.err}

			got, err := getHomeId(context.Background(), f, discardLogger)
			assert.Equal(t, tt.want.homeId, got)
			tt.want.err(t, err)
		})
	}
}

var _ publisher.WeatherGetter = fakeTadoGetter{}

type fakeTadoGetter struct{}

func (f fakeTadoGetter) GetWeatherWithResponse(_ context.Context, _ tado.HomeId, _ ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error) {
	return &tado.GetWeatherResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: http.NoBody},
		JSON200: &tado.Weather{
			SolarIntensity: &tado.PercentageDataPoint{Percentage: pointer(float32(75))},
			WeatherState:   &tado.WeatherStateDataPoint{Value: pointer(tado.SUN)},
		},
	}, nil
}
