package export

import (
	"errors"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/pubsub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

var (
	Cmd = cobra.Command{
		Use:   "export",
		Short: "collect SolarEdge data and export to Prometheus",
		RunE:  run,
	}
)

func run(cmd *cobra.Command, _ []string) error {
	logger := charmer.GetLogger(cmd)

	logger.Info("starting solaredge exporter", "version", cmd.Root().Version)
	defer logger.Info("stopping solaredge exporter")

	go func() {
		err := http.ListenAndServe(viper.GetString("prometheus.addr"), promhttp.Handler())
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "exporter", ConstLabels: prometheus.Labels{"application": "solaredge"}})
	prometheus.MustRegister(solarEdgeMetrics)

	httpClient := http.Client{
		Timeout:   5 * time.Second,
		Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
	}

	poller := scraper.Poller{
		Client:    solaredge.New(viper.GetString("polling.token"), &httpClient),
		Interval:  viper.GetDuration("polling.interval"),
		Logger:    logger.With("component", "poller"),
		Publisher: pubsub.Publisher[solaredge.Update]{},
	}

	exportMetrics := scraper.NewMetrics()
	prometheus.MustRegister(exportMetrics)

	exporter := scraper.Exporter{
		Poller:  &poller,
		Metrics: exportMetrics,
		Logger:  logger.With("component", "exporter"),
	}

	var group errgroup.Group
	group.Go(func() error { return poller.Run(cmd.Context()) })
	group.Go(func() error { return exporter.Run(cmd.Context()) })

	return group.Wait()
}
