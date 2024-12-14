package cmd

import (
	"context"
	solaredge2 "github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/poller"
	"github.com/clambin/solaredge-monitor/internal/poller/solaredge"
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
	p := fakeUpdater{Update: solaredge.Update{{
		ID:   1,
		Name: "my home",
		PowerOverview: solaredge2.PowerOverview{
			LastUpdateTime: solaredge2.Time(time.Date(2024, time.December, 12, 12, 0, 0, 0, time.UTC)),
			LifeTimeData:   solaredge2.EnergyOverview{Energy: 1000},
			LastYearData:   solaredge2.EnergyOverview{Energy: 100},
			LastMonthData:  solaredge2.EnergyOverview{Energy: 10},
			LastDayData:    solaredge2.EnergyOverview{Energy: 1},
			CurrentPower:   solaredge2.CurrentPower{Power: 500},
		},
	}}}
	v := getViperFromViper(viper.GetViper())
	v.Set("polling.interval", time.Second)
	r := prometheus.NewPedanticRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		errCh <- runExport(ctx, "dev", v, r, &p, discardLogger)
	}()

	var metricNames = []string{
		"solaredge_current_power",
		"solaredge_day_energy",
		"solaredge_month_energy",
		"solaredge_year_energy",
	}

	assert.Eventually(t, func() bool {
		count, err := testutil.GatherAndCount(r, metricNames...)
		return err == nil && count == len(metricNames)
	}, 10*time.Second, time.Millisecond)

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

var _ poller.Updater[solaredge.Update] = fakeUpdater{}

type fakeUpdater struct {
	solaredge.Update
}

func (f fakeUpdater) GetUpdate(_ context.Context) (solaredge.Update, error) {
	return f.Update, nil
}
