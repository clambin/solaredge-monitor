package cmd

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/solaredge-monitor/internal/exporter"
	"github.com/clambin/solaredge-monitor/internal/poller"
	solaredge2 "github.com/clambin/solaredge-monitor/internal/poller/solaredge"
	tado2 "github.com/clambin/solaredge-monitor/internal/poller/tado"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/clambin/tado/v2"
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
			solarEdgeClient := newSolarEdgeClient("scraper", prometheus.DefaultRegisterer, viper.GetViper())
			tadoClient, err := newTadoClient(ctx, prometheus.DefaultRegisterer, viper.GetViper())
			if err != nil {
				return fmt.Errorf("tado: %w", err)
			}
			homeId, err := getHomeId(ctx, tadoClient, logger)
			if err != nil {
				return fmt.Errorf("failed to list Tado Homes: %w", err)
			}
			return runScrape(ctx, cmd.Root().Version, viper.GetViper(), prometheus.DefaultRegisterer, solaredge2.Client{SolarEdge: solarEdgeClient}, tado2.Client{TadoClient: tadoClient}, homeId, logger)
		},
	}
)

func runScrape(
	ctx context.Context,
	version string,
	v *viper.Viper,
	r prometheus.Registerer,
	solarEdgeUpdater poller.Updater[solaredge2.Update],
	tadoUpdater poller.Updater[*tado.Weather],
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

	solarEdgePoller := poller.Poller[solaredge2.Update]{
		Updater:  solarEdgeUpdater,
		Interval: v.GetDuration("polling.interval"),
		Logger:   logger.With("poller", "solaredge"),
	}

	tadoPoller := poller.Poller[*tado.Weather]{
		Updater:  tadoUpdater,
		Interval: v.GetDuration("polling.interval"),
		Logger:   logger.With("poller", "tado"),
	}

	writer := scraper.Writer{
		Store:     repo,
		SolarEdge: &solarEdgePoller,
		Tado:      &tadoPoller,
		HomeId:    homeId,
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
