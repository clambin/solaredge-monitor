package cmd

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"
)

var discardLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

func Test_runExport(t *testing.T) {
	v := getViperFromViper(viper.GetViper())
	p := fakePoller{ch: make(chan solaredge.Update)}
	r := prometheus.NewPedanticRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- runExport(ctx, "dev", v, r, &p, discardLogger)
	}()

	go feed(ctx, p.ch, 5, 500*time.Millisecond)

	var metricNames = []string{
		"solaredge_current_power",
		"solaredge_day_energy",
		"solaredge_month_energy",
		"solaredge_year_energy",
	}

	assert.Eventually(t, func() bool {
		count, err := testutil.GatherAndCount(r, metricNames...)
		return err == nil && count == len(metricNames)
	}, time.Second, time.Millisecond)

	assert.NoError(t, testutil.GatherAndCompare(r, strings.NewReader(`
# HELP solaredge_current_power current power in Watt
# TYPE solaredge_current_power gauge
solaredge_current_power{site="my home"} 500

# HELP solaredge_day_energy Today's produced energy in WattHours
# TYPE solaredge_day_energy gauge
solaredge_day_energy{site="my home"} 1

# HELP solaredge_month_energy This month's produced energy in WattHours
# TYPE solaredge_month_energy gauge
solaredge_month_energy{site="my home"} 10
# HELP solaredge_year_energy This year's produced energy in WattHours
# TYPE solaredge_year_energy gauge
solaredge_year_energy{site="my home"} 100
`), metricNames...))

	cancel()
	assert.NoError(t, <-errCh)
}

var _ Poller = fakePoller{}

type fakePoller struct {
	ch chan solaredge.Update
}

func (f fakePoller) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (f fakePoller) Subscribe() chan solaredge.Update {
	return f.ch
}

func (f fakePoller) Unsubscribe(_ chan solaredge.Update) {
}
