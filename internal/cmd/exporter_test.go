package cmd

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func Test_runExport(t *testing.T) {
	p := fakeUpdater{SolarEdgeUpdate: publisher.SolarEdgeUpdate{{
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

var _ publisher.Updater[publisher.SolarEdgeUpdate] = fakeUpdater{}

type fakeUpdater struct {
	publisher.SolarEdgeUpdate
}

func (f fakeUpdater) GetUpdate(_ context.Context) (publisher.SolarEdgeUpdate, error) {
	return f.SolarEdgeUpdate, nil
}
