package cmd

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	tadov2 "github.com/clambin/tado/v2"
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
	ctx, cancel := context.WithCancel(context.Background())
	store, port, err := testutils.NewTestPostgresDB(ctx, "solaredge", "username", "password")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(store))
	})
	v := getViperFromViper(viper.GetViper())
	initViperDB(v, port)
	v.Set("scrape.interval", time.Second)
	p := fakePoller{ch: make(chan solaredge.Update)}
	r := prometheus.NewPedanticRegistry()
	c := fakeTadoGetter{}

	errCh := make(chan error)
	go func() {
		errCh <- runScrape(ctx, "dev", v, r, &p, c, 1, discardLogger)
	}()

	go feed(ctx, p.ch, 5, 500*time.Millisecond)

	dbc, err := repository.NewPostgresDB("localhost", port, "solaredge", "username", "password")
	require.NoError(t, err)
	assert.Eventually(t, func() bool {
		rows, err := dbc.Get(time.Time{}, time.Time{})
		return err == nil && len(rows) > 0

	}, 10*time.Second, time.Millisecond)

	cancel()
	assert.NoError(t, <-errCh)
}

var _ scraper.TadoGetter = fakeTadoGetter{}

type fakeTadoGetter struct{}

func (f fakeTadoGetter) GetWeatherWithResponse(_ context.Context, _ tadov2.HomeId, _ ...tadov2.RequestEditorFn) (*tadov2.GetWeatherResponse, error) {
	return &tadov2.GetWeatherResponse{
		HTTPResponse: &http.Response{StatusCode: http.StatusOK, Body: http.NoBody},
		JSON200: &tadov2.Weather{
			SolarIntensity: &tadov2.PercentageDataPoint{Percentage: varP(float32(75))},
			WeatherState:   &tadov2.WeatherStateDataPoint{Value: varP(tadov2.SUN)},
		},
	}, nil
}
