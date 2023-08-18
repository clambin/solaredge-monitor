package collector

import (
	"context"
	"github.com/clambin/solaredge-monitor/collector/solaredgeScraper"
	"github.com/clambin/solaredge-monitor/collector/tadoScraper"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
	"sync"
	"time"
)

type Collector struct {
	TadoScraper      Scraper[tadoScraper.Info]
	SolarEdgeScraper Scraper[solaredgeScraper.Info]
	ScrapeInterval   time.Duration
	CollectInterval  time.Duration
	Logger           *slog.Logger

	//temperature  Averager
	intensity Averager
	power     Averager
	weather   WordCounter
	store.DB
}

type Scraper[T any] interface {
	Run(ctx context.Context, duration time.Duration, info chan<- T)
}

var (
	collectorSamples = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName("solaredgescraper", "collector", "summary"),
		Help: "Number of samples in a collection",
	}, []string{"collector"})
)

func (c *Collector) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	tadoInfo := make(chan tadoScraper.Info, 1)
	go func() {
		defer wg.Done()
		c.TadoScraper.Run(ctx, c.ScrapeInterval, tadoInfo)
	}()

	solarEdgeInfo := make(chan solaredgeScraper.Info, 1)
	go func() {
		defer wg.Done()
		c.SolarEdgeScraper.Run(ctx, c.ScrapeInterval, solarEdgeInfo)
	}()

	ticker := time.NewTicker(c.CollectInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			c.collect()
			return nil
		case info := <-tadoInfo:
			c.intensity.Add(info.SolarIntensity)
			//c.temperature.Add(info.Temperature)
			c.weather.Add(info.Weather)
		case info := <-solarEdgeInfo:
			c.power.Add(info.Power)
		case <-ticker.C:
			c.collect()
		}
	}
}

func (c *Collector) collect() {
	if c.power.Count == 0 || c.intensity.Count == 0 {
		c.Logger.Warn("partial data collection. skipping")
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
		c.Logger.Debug("no solar power activity. skipping measurement")
		return
	}

	if err := c.Store(measurement); err != nil {
		c.Logger.Error("failed to store metrics", "err", err)
		return
	}

	c.Logger.Info("new entry", slog.Group("measurement",
		slog.Float64("power", measurement.Power),
		slog.Float64("intensity", measurement.Intensity),
		slog.String("weather", measurement.Weather),
	))
}
