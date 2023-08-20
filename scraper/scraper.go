package scraper

import (
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log/slog"
	"time"
)

var (
	collectorSamples = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: prometheus.BuildFQName("solaredgescraper", "collector", "summary"),
		Help: "Number of samples in a collection",
	}, []string{"scraper"})
)

func New(powerGetter PowerGetter, weatherGetter WeatherGetter, db store.DB, l *slog.Logger, scrapeInterval, collectInterval time.Duration) *taskmanager.Manager {
	powerScraper := &PowerScraper{PowerGetter: powerGetter}
	weatherScraper := &WeatherScraper{WeatherGetter: weatherGetter}

	return taskmanager.New(
		&daemon{
			Poller:   powerScraper,
			Interval: scrapeInterval,
			Logger:   l.With("component", "powerScraper"),
		},
		&daemon{
			Poller:   weatherScraper,
			Interval: scrapeInterval,
			Logger:   l.With("component", "weatherScraper"),
		},
		&Collector{
			Interval:       collectInterval,
			DB:             db,
			Logger:         l.With("component", "collector"),
			WeatherScraper: weatherScraper,
			PowerScraper:   powerScraper,
		},
	)
}
