package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
)

var (
	scrapeCmd = cobra.Command{
		Use:   "scrape",
		Short: "collect SolarEdge data and export to Prometheus & Postgres",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := charmer.GetLogger(cmd)
			poller := newPoller(prometheus.DefaultRegisterer, viper.GetViper(), logger.With("component", "poller"))
			tadoClient, err := newTadoClient(ctx)
			if err != nil {
				return fmt.Errorf("tado: %w", err)
			}
			homeId, err := getHomeId(ctx, tadoClient, logger)
			if err != nil {
				return fmt.Errorf("failed to list Tado Homes: %w", err)
			}
			return runScrape(ctx, cmd.Root().Version, viper.GetViper(), prometheus.DefaultRegisterer, poller, tadoClient, homeId, logger)
		},
	}
)

func runScrape(
	ctx context.Context,
	version string,
	v *viper.Viper,
	r prometheus.Registerer,
	poller Poller,
	tadoClient scraper.TadoGetter,
	homeId tado.HomeId,
	logger *slog.Logger,
) error {
	logger.Info("starting solaredge scraper", "version", version)
	defer logger.Info("stopping solaredge scraper")

	repo, err := repository.NewPostgresDB(
		v.GetString("database.host"),
		v.GetInt("database.port"),
		v.GetString("database.database"),
		v.GetString("database.username"),
		v.GetString("database.password"),
	)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	logger.Debug("connected to database")

	writer := scraper.Writer{
		Store:      repo,
		TadoGetter: tadoClient,
		HomeId:     homeId,
		Poller:     poller,
		Interval:   v.GetDuration("scrape.interval"),
		Logger:     logger.With("component", "writer"),
	}

	exportMetrics := scraper.NewMetrics()
	r.MustRegister(exportMetrics)

	exporter := scraper.Exporter{
		Poller:  poller,
		Metrics: exportMetrics,
		Logger:  logger.With("component", "exporter"),
	}

	var group errgroup.Group
	group.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: v.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	group.Go(func() error { return exporter.Run(ctx) })
	group.Go(func() error { return writer.Run(ctx) })
	group.Go(func() error { return poller.Run(ctx) })

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
