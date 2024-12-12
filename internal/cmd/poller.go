package cmd

import (
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/go-common/pubsub"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
)

func newPoller(r prometheus.Registerer, v *viper.Viper, l *slog.Logger) *scraper.Poller {
	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "exporter", ConstLabels: prometheus.Labels{"application": "solaredge"}})
	r.MustRegister(solarEdgeMetrics)

	return &scraper.Poller{
		Client: solaredge.New(v.GetString("polling.token"), &http.Client{
			Timeout:   5 * time.Second,
			Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
		}),
		Interval:  v.GetDuration("polling.interval"),
		Logger:    l,
		Publisher: pubsub.Publisher[solaredge.Update]{},
	}
}
