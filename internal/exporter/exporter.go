package exporter

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/publisher/solaredge"
	"github.com/prometheus/client_golang/prometheus"
	"log/slog"
)

type Exporter struct {
	SolarEdge Publisher[solaredge.Update]
	Metrics   *Metrics
	Logger    *slog.Logger
}

type Publisher[T any] interface {
	Subscribe() chan T
	Unsubscribe(chan T)
}

func (e Exporter) Run(ctx context.Context) error {
	ch := e.SolarEdge.Subscribe()
	defer e.SolarEdge.Unsubscribe(ch)

	e.Logger.Debug("starting exporter")
	defer e.Logger.Debug("stopped exporter")

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-ch:
			e.export(update)
		}
	}
}

func (e Exporter) export(update solaredge.Update) {
	e.Logger.Debug("exporting update")
	for site := range update {
		e.Metrics.currentPower.WithLabelValues(update[site].Name).Set(update[site].PowerOverview.CurrentPower.Power)
		e.Metrics.dayEnergy.WithLabelValues(update[site].Name).Set(update[site].PowerOverview.LastDayData.Energy)
		e.Metrics.monthEnergy.WithLabelValues(update[site].Name).Set(update[site].PowerOverview.LastMonthData.Energy)
		e.Metrics.yearEnergy.WithLabelValues(update[site].Name).Set(update[site].PowerOverview.LastYearData.Energy)

		for inverter := range update[site].InverterUpdates {
			e.Metrics.inverterTemperature.WithLabelValues(update[site].Name, update[site].InverterUpdates[inverter].Name).Set(update[site].InverterUpdates[inverter].Telemetry.Temperature)
			e.Metrics.inverterACVoltage.WithLabelValues(update[site].Name, update[site].InverterUpdates[inverter].Name).Set(update[site].InverterUpdates[inverter].Telemetry.L1Data.AcVoltage)
			e.Metrics.inverterACCurrent.WithLabelValues(update[site].Name, update[site].InverterUpdates[inverter].Name).Set(update[site].InverterUpdates[inverter].Telemetry.L1Data.AcCurrent)
			e.Metrics.inverterDCVoltage.WithLabelValues(update[site].Name, update[site].InverterUpdates[inverter].Name).Set(update[site].InverterUpdates[inverter].Telemetry.DcVoltage)
			e.Metrics.inverterPowerLimit.WithLabelValues(update[site].Name, update[site].InverterUpdates[inverter].Name).Set(update[site].InverterUpdates[inverter].Telemetry.PowerLimit)
		}
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////

var _ prometheus.Collector = Metrics{}

type Metrics struct {
	currentPower        *prometheus.GaugeVec
	dayEnergy           *prometheus.GaugeVec
	monthEnergy         *prometheus.GaugeVec
	yearEnergy          *prometheus.GaugeVec
	inverterTemperature *prometheus.GaugeVec
	inverterACVoltage   *prometheus.GaugeVec
	inverterACCurrent   *prometheus.GaugeVec
	inverterDCVoltage   *prometheus.GaugeVec
	inverterPowerLimit  *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		currentPower: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "", "current_power"),
			Help: "current power in Watt",
		}, []string{"site"}),
		dayEnergy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "", "day_energy"),
			Help: "Today's produced energy in WattHours",
		}, []string{"site"}),
		monthEnergy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "", "month_energy"),
			Help: "This month's produced energy in WattHours",
		}, []string{"site"}),
		yearEnergy: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "", "year_energy"),
			Help: "This year's produced energy in WattHours",
		}, []string{"site"}),
		inverterTemperature: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "inverter", "temperature"),
			Help: "Temperature reported by the inverter(s)",
		}, []string{"site", "inverter"}),
		inverterACVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "inverter", "ac_voltage"),
			Help: "AC voltage reported by the inverter(s)",
		}, []string{"site", "inverter"}),
		inverterACCurrent: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "inverter", "ac_current"),
			Help: "AC current reported by the inverter(s)",
		}, []string{"site", "inverter"}),
		inverterDCVoltage: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "inverter", "dc_voltage"),
			Help: "DC voltage reported by the inverter(s)",
		}, []string{"site", "inverter"}),
		inverterPowerLimit: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName("solaredge", "inverter", "power_limit"),
			Help: "Power limit reported by the inverter(s)",
		}, []string{"site", "inverter"}),
	}
}

func (m Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.currentPower.Describe(ch)
	m.dayEnergy.Describe(ch)
	m.monthEnergy.Describe(ch)
	m.yearEnergy.Describe(ch)
	m.inverterTemperature.Describe(ch)
	m.inverterACVoltage.Describe(ch)
	m.inverterACCurrent.Describe(ch)
	m.inverterDCVoltage.Describe(ch)
	m.inverterPowerLimit.Describe(ch)
}

func (m Metrics) Collect(ch chan<- prometheus.Metric) {
	m.currentPower.Collect(ch)
	m.dayEnergy.Collect(ch)
	m.monthEnergy.Collect(ch)
	m.yearEnergy.Collect(ch)
	m.inverterTemperature.Collect(ch)
	m.inverterACVoltage.Collect(ch)
	m.inverterACCurrent.Collect(ch)
	m.inverterDCVoltage.Collect(ch)
	m.inverterPowerLimit.Collect(ch)
}
