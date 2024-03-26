package scrape

import (
	"errors"
	"fmt"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	solaredge2 "github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/pubsub"
	"github.com/clambin/tado"
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
		Use:   "scrape",
		Short: "collect SolarEdge data and export to Prometheus & Postgres",
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

	repo, err := repository.NewPostgresDB(
		viper.GetString("database.host"),
		viper.GetInt("database.port"),
		viper.GetString("database.database"),
		viper.GetString("database.username"),
		viper.GetString("database.password"),
	)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	logger.Debug("connected to database")

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

	tadoMetrics := metrics.NewRequestSummaryMetrics("solaredge", "scraper", map[string]string{"application": "tado"})
	prometheus.MustRegister(tadoMetrics)

	tadoClient, err := tado.New(
		viper.GetString("tado.username"),
		viper.GetString("tado.password"),
		viper.GetString("tado.secret"),
	)
	if err != nil {
		return fmt.Errorf("tado: %w", err)
	}

	origTP := tadoClient.HTTPClient.Transport
	tadoClient.HTTPClient.Transport = roundtripper.New(
		roundtripper.WithRequestMetrics(tadoMetrics),
		roundtripper.WithRoundTripper(origTP),
	)

	writer := scraper.Writer{
		Store:      repo,
		TadoClient: tadoClient,
		Poller:     &poller,
		Interval:   viper.GetDuration("scrape.collection"),
		Logger:     logger.With("component", "writer"),
	}

	var group errgroup.Group
	group.Go(func() error { return poller.Run(cmd.Context()) })
	group.Go(func() error { return exporter.Run(cmd.Context()) })
	group.Go(func() error { return writer.Run(cmd.Context()) })

	return group.Wait()
}
