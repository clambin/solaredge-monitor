package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/roundtripper"
	"github.com/clambin/solaredge-monitor/internal/exporter"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/tado/v2"
	"github.com/clambin/tado/v2/tools"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)

var (
	scrapeCmd = cobra.Command{
		Use:   "scrape",
		Short: "collect SolarEdge data and export to Prometheus & Postgres",
		PreRun: func(cmd *cobra.Command, args []string) {
			charmer.SetJSONLogger(cmd, viper.GetBool("debug"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := charmer.GetLogger(cmd)
			solarEdgeClient := newSolarEdgeClient("scraper", prometheus.DefaultRegisterer, viper.GetString("solaredge.token"))
			tadoClient, err := newTadoClient(ctx, prometheus.DefaultRegisterer, viper.GetString("tado.token.path"), viper.GetString("tado.token.passphrase"))
			if err != nil {
				return fmt.Errorf("tado: %w", err)
			}
			homeId, err := getHomeId(ctx, tadoClient, logger)
			if err != nil {
				return fmt.Errorf("failed to list Tado Homes: %w", err)
			}
			return runScrape(
				ctx,
				cmd.Root().Version,
				viper.GetViper(),
				prometheus.DefaultRegisterer,
				publisher.SolarEdgeUpdater{SolarEdgeClient: &solarEdgeClient},
				publisher.TadoUpdater{Client: tadoClient, HomeId: homeId},
				logger,
			)
		},
	}
)

func runScrape(
	ctx context.Context,
	version string,
	v *viper.Viper,
	r prometheus.Registerer,
	solarEdgeUpdater publisher.Updater[publisher.SolarEdgeUpdate],
	tadoUpdater publisher.Updater[*tado.Weather],
	logger *slog.Logger,
) error {
	logger.Info("starting solaredge scraper", "version", version)
	defer logger.Info("stopping solaredge scraper")

	if pprofAddr := v.GetString("pprof"); pprofAddr != "" {
		go func() {
			logger.Debug("starting pprof", "addr", pprofAddr)
			_ = http.ListenAndServe(pprofAddr, nil)
		}()
	}

	repo, err := repository.NewPostgresDB(v.GetString("database.url"))
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	logger.Debug("connected to database")

	solarEdgePoller := publisher.Publisher[publisher.SolarEdgeUpdate]{
		Updater:  solarEdgeUpdater,
		Interval: v.GetDuration("polling.interval"),
		Logger:   logger.With("publisher", "solaredge"),
	}

	tadoPoller := publisher.Publisher[*tado.Weather]{
		Updater:  tadoUpdater,
		Interval: v.GetDuration("polling.interval"),
		Logger:   logger.With("publisher", "tado"),
	}

	writer := scraper.Writer{
		Store:     repo,
		SolarEdge: &solarEdgePoller,
		Tado:      &tadoPoller,
		Interval:  v.GetDuration("scrape.interval"),
		Logger:    logger.With("component", "writer"),
	}

	exportMetrics := exporter.NewMetrics()
	r.MustRegister(exportMetrics)

	exp := exporter.Exporter{
		SolarEdge: &solarEdgePoller,
		Metrics:   exportMetrics,
		Logger:    logger.With("component", "exporter"),
	}

	var group errgroup.Group
	group.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: v.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	group.Go(func() error { return writer.Run(ctx) })
	group.Go(func() error { return exp.Run(ctx) })
	group.Go(func() error { return solarEdgePoller.Run(ctx) })
	group.Go(func() error { return tadoPoller.Run(ctx) })

	return group.Wait()
}

func newTadoClient(ctx context.Context, r prometheus.Registerer, path, passphrase string) (*tado.ClientWithResponses, error) {
	tadoMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "scraper", ConstLabels: prometheus.Labels{"application": "tado"}})
	r.MustRegister(tadoMetrics)

	tadoHttpClient, err := tado.NewOAuth2Client(ctx, path, passphrase, func(response *oauth2.DeviceAuthResponse) {
		fmt.Printf("No token found. Visit %s and log in ...\n", response.VerificationURIComplete)
	})
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

func getHomeId(ctx context.Context, client tools.TadoClient, logger *slog.Logger) (tado.HomeId, error) {
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
