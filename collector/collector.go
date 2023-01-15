package collector

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector/solaredgescraper"
	"github.com/clambin/solaredge-monitor/collector/tadoscraper"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/tado"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/slog"
	"sync"
	"time"
)

type Collector struct {
	TadoAPI      tado.API
	SolarEdgeAPI solaredge.API
	//temperature  Averager
	intensity Averager
	power     Averager
	weather   WordCounter
	store.DB
}

var (
	collectorSamples = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName("solaredgescraper", "collector", "summary"),
		Help: "Number of samples in a collection",
	}, []string{"collector"})
)

func (c *Collector) Run(ctx context.Context, scrapeInterval time.Duration, collectInterval time.Duration) {
	var wg sync.WaitGroup
	wg.Add(2)

	tadoInfo := make(chan tadoscraper.Info, 1)
	go func() {
		f := tadoscraper.Fetcher{API: c.TadoAPI}
		f.Run(ctx, scrapeInterval, tadoInfo)
		wg.Done()
	}()

	solaredgeInfo := make(chan solaredgescraper.Info, 1)
	go func() {
		f := solaredgescraper.Fetcher{API: c.SolarEdgeAPI}
		f.Run(ctx, scrapeInterval, solaredgeInfo)
		wg.Done()
	}()

	ticker := time.NewTicker(collectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			c.collect()
			return
		case info := <-tadoInfo:
			c.intensity.Add(info.SolarIntensity)
			//c.temperature.Add(info.Temperature)
			c.weather.Add(info.Weather)
		case info := <-solaredgeInfo:
			c.power.Add(info.Power)
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *Collector) collect() {
	if c.power.Count == 0 || c.intensity.Count == 0 {
		slog.Warn("partial data collection. skipping")
		return
	}

	collectorSamples.WithLabelValues("solaredgescraper").Set(float64(c.power.Count))
	collectorSamples.WithLabelValues("tadoscraper").Set(float64(c.intensity.Count))

	measurement := store.Measurement{
		Timestamp: time.Now(),
		Power:     c.power.Average(),
		Intensity: c.intensity.Average(),
		Weather:   c.weather.GetMostUsed(),
	}

	if measurement.Power == 0 && measurement.Intensity == 0 {
		slog.Debug("no solar power activity. skipping measurement")
		return
	}

	if err := c.Store(measurement); err != nil {
		slog.Error("failed to store metrics", err)
		return
	}

	slog.Info("new entry", slog.Group("measurement",
		slog.Float64("power", measurement.Power),
		slog.Float64("intensity", measurement.Intensity),
		slog.String("weather", measurement.Weather),
	))
}
