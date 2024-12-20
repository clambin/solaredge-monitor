package exporter

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge-monitor/internal/testutils"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestExporter(t *testing.T) {
	p := testutils.FakePublisher[publisher.SolarEdgeUpdate]{Ch: make(chan publisher.SolarEdgeUpdate)}

	metrics := NewMetrics()
	exporter := Exporter{
		SolarEdge: &p,
		Metrics:   metrics,
		Logger:    slog.Default(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() { _ = exporter.Run(ctx) }()
	p.Ch <- testutils.TestUpdate

	require.Eventually(t, func() bool {
		return testutil.CollectAndCount(metrics) >= 9
	}, 10*time.Second, time.Millisecond)

	assert.NoError(t, testutil.CollectAndCompare(metrics, strings.NewReader(`
# HELP solaredge_current_power current power in Watt
# TYPE solaredge_current_power gauge
solaredge_current_power{site="foo"} 3000

# HELP solaredge_day_energy Today's produced energy in WattHours
# TYPE solaredge_day_energy gauge
solaredge_day_energy{site="foo"} 10

# HELP solaredge_inverter_ac_current AC current reported by the inverter(s)
# TYPE solaredge_inverter_ac_current gauge
solaredge_inverter_ac_current{inverter="inv1",site="foo"} 10

# HELP solaredge_inverter_ac_voltage AC voltage reported by the inverter(s)
# TYPE solaredge_inverter_ac_voltage gauge
solaredge_inverter_ac_voltage{inverter="inv1",site="foo"} 240

# HELP solaredge_inverter_dc_voltage DC voltage reported by the inverter(s)
# TYPE solaredge_inverter_dc_voltage gauge
solaredge_inverter_dc_voltage{inverter="inv1",site="foo"} 400

# HELP solaredge_inverter_power_limit Power limit reported by the inverter(s)
# TYPE solaredge_inverter_power_limit gauge
solaredge_inverter_power_limit{inverter="inv1",site="foo"} 1

# HELP solaredge_inverter_temperature Temperature reported by the inverter(s)
# TYPE solaredge_inverter_temperature gauge
solaredge_inverter_temperature{inverter="inv1",site="foo"} 40

# HELP solaredge_month_energy This month's produced energy in WattHours
# TYPE solaredge_month_energy gauge
solaredge_month_energy{site="foo"} 100

# HELP solaredge_year_energy This year's produced energy in WattHours
# TYPE solaredge_year_energy gauge
solaredge_year_energy{site="foo"} 1000
`)))
}
