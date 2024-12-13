package cmd

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
)

var (
	exportCmd = cobra.Command{
		Use:   "export",
		Short: "collect SolarEdge data and export to Prometheus",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := charmer.GetLogger(cmd)
			poller := newPoller(prometheus.DefaultRegisterer, viper.GetViper(), logger.With("component", "poller"))
			return runExport(ctx, cmd.Root().Version, viper.GetViper(), prometheus.DefaultRegisterer, poller, logger)
		},
	}
)

func runExport(
	ctx context.Context,
	version string,
	v *viper.Viper,
	r prometheus.Registerer,
	poller Poller,
	logger *slog.Logger,
) error {
	logger.Info("starting solaredge exporter", "version", version)
	defer logger.Info("stopping solaredge exporter")

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
	group.Go(func() error { return poller.Run(ctx) })
	group.Go(func() error { return exporter.Run(ctx) })

	return group.Wait()
}
