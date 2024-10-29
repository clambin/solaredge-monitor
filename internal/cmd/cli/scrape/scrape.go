package scrape

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/roundtripper"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/pubsub"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
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
	logger := charmer.GetLogger(cmd)

	logger.Info("starting solaredge scraper", "version", cmd.Root().Version)
	defer logger.Info("stopping solaredge scraper")

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

	solarEdgeMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "scraper", ConstLabels: prometheus.Labels{"application": "solaredge"}})
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

	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	tadoClient, err := newTadoClient(ctx)
	if err != nil {
		return fmt.Errorf("tado: %w", err)
	}

	homeId, err := getHomeId(ctx, tadoClient, logger)
	if err != nil {
		return fmt.Errorf("failed to list Tado Homes: %w", err)
	}

	writer := scraper.Writer{
		Store:      repo,
		TadoGetter: tadoClient,
		HomeId:     homeId,
		Poller:     &poller,
		Interval:   viper.GetDuration("scrape.interval"),
		Logger:     logger.With("component", "writer"),
	}

	exportMetrics := scraper.NewMetrics()
	prometheus.MustRegister(exportMetrics)

	exporter := scraper.Exporter{
		Poller:  &poller,
		Metrics: exportMetrics,
		Logger:  logger.With("component", "exporter"),
	}

	var group errgroup.Group
	group.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: viper.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	group.Go(func() error { return poller.Run(ctx) })
	group.Go(func() error { return exporter.Run(ctx) })
	group.Go(func() error { return writer.Run(ctx) })

	return group.Wait()
}

func newTadoClient(ctx context.Context) (*tado.ClientWithResponses, error) {
	tadoMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "scraper", ConstLabels: prometheus.Labels{"application": "tado"}})
	prometheus.MustRegister(tadoMetrics)

	tadoHttpClient, err := tado.NewOAuth2Client(ctx, viper.GetString("tado.username"), viper.GetString("tado.password"))
	if err != nil {
		return nil, err
	}
	origTP := tadoHttpClient.Transport
	tadoHttpClient.Transport = roundtripper.New(
		roundtripper.WithRequestMetrics(tadoMetrics),
		roundtripper.WithRoundTripper(origTP),
	)
	return tado.NewClientWithResponses(tado.ServerURL, tado.WithHTTPClient(tadoHttpClient))
}

func getHomeId(ctx context.Context, client *tado.ClientWithResponses, logger *slog.Logger) (tado.HomeId, error) {
	homes, err := tools.GetHomes(ctx, client)
	if err != nil {
		return 0, err
	}
	if len(homes) == 0 {
		return 0, errors.New("no Tado Homes found")
	}
	homeId := *homes[0].Id
	if len(homes) > 1 {
		logger.Warn("Tado account has more than one home registered. Using first one", "homeId", homeId)
	}
	return homeId, nil
}
