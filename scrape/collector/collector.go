package collector

import (
	"context"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	"time"
)

type Collector struct {
	SolarEdge scraper.Summarizer
	Tado      scraper.Summarizer
	store.DB
}

var (
	collectorSamples = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName("solaredge", "collector", "summary"),
		Help: "Number of samples in a collection",
	}, []string{"collector"})
)

func (c *Collector) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
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
}

func (c *Collector) collect() {
	if c.SolarEdge.Count() == 0 || c.Tado.Count() == 0 {
		log.Warning("partial data collection. skipping")
		return
	}

	collectorSamples.WithLabelValues("solaredge").Set(float64(c.SolarEdge.Count()))
	collectorSamples.WithLabelValues("tado").Set(float64(c.Tado.Count()))

	power := c.SolarEdge.Summarize()
	intensity := c.Tado.Summarize()

	ts := power.Timestamp
	if intensity.Timestamp.Before(ts) {
		ts = intensity.Timestamp
	}

	measurement := store.Measurement{
		Timestamp: ts,
		Power:     power.Value,
		Intensity: intensity.Value,
	}

	if err := c.Store(measurement); err == nil {
		log.WithField("measurement", measurement).Info("new entry")
	} else {
		log.WithError(err).Warning("failed to store metrics")
	}
}
