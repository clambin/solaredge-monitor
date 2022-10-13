package collector

import (
	"context"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Collector struct {
	SolarEdge scraper.Summarizer
	Tado      scraper.Summarizer
	store.DB
	Interval time.Duration
}

var (
	collectorSamples = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName("solaredge", "collector", "summary"),
		Help: "Number of samples in a collection",
	}, []string{"collector"})
)

func (c *Collector) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		c.SolarEdge.Run(ctx)
		wg.Done()
	}()
	go func() {
		c.Tado.Run(ctx)
		wg.Done()
	}()

	ticker := time.NewTicker(c.Interval)
	for running := true; running; {
		select {
		case <-ctx.Done():
			running = false
		case <-ticker.C:
			c.collect()
		}
	}
	ticker.Stop()
	c.collect()

	wg.Wait()
}

func (c *Collector) collect() {
	if c.SolarEdge.Count() == 0 || c.Tado.Count() == 0 {
		log.Warning("partial data collection. skipping")
		return
	}

	collectorSamples.WithLabelValues("solaredge").Set(float64(c.SolarEdge.Count()))
	collectorSamples.WithLabelValues("tado").Set(float64(c.Tado.Count()))

	measurement := store.Measurement{
		Timestamp: time.Now(),
		Power:     c.SolarEdge.Summarize().Value,
		Intensity: c.Tado.Summarize().Value,
	}

	if err := c.Store(measurement); err == nil {
		log.WithFields(log.Fields{
			"power":     measurement.Power,
			"intensity": measurement.Intensity,
		}).Info("new entry")
	} else {
		log.WithError(err).Warning("failed to store metrics")
	}
}
