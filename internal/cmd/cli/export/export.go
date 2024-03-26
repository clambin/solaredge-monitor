package export

import (
	"errors"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	solaredge2 "github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/pubsub"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	"os"
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
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &opts))

	go func() {
		err := http.ListenAndServe(viper.GetString("prometheus.addr"), promhttp.Handler())
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	solarEdgeMetrics := metrics.NewRequestSummaryMetrics("solaredge", "scraper", map[string]string{"application": "solaredge"})
	prometheus.MustRegister(solarEdgeMetrics)

	httpClient := http.Client{
		Timeout:   5 * time.Second,
		Transport: roundtripper.New(roundtripper.WithRequestMetrics(solarEdgeMetrics)),
	}

	poller := scraper.Poller{
		Client:    solaredge2.New(viper.GetString("solaredge.token"), &httpClient),
		Interval:  viper.GetDuration("scrape.polling"),
		Logger:    logger.With("component", "poller"),
		Publisher: pubsub.Publisher[solaredge2.Update]{},
	}

	exportMetrics := scraper.NewMetrics()
	prometheus.MustRegister(exportMetrics)

	exporter := scraper.Exporter{
		Poller:  &poller,
		Metrics: exportMetrics,
	}

	var group errgroup.Group
	group.Go(func() error { return poller.Run(cmd.Context()) })
	group.Go(func() error { return exporter.Run(cmd.Context()) })

	return group.Wait()
}
