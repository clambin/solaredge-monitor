package cmd

import (
	"context"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/solaredge-monitor/internal/exporter"
	"github.com/clambin/solaredge-monitor/internal/publisher"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)

var (
	exportCmd = cobra.Command{
		Use:   "export",
		Short: "collect SolarEdge data and export to Prometheus",
		PreRun: func(cmd *cobra.Command, args []string) {
			charmer.SetJSONLogger(cmd, viper.GetBool("debug"))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			logger := charmer.GetLogger(cmd)
			solarEdgeClient := newSolarEdgeClient("exporter", prometheus.DefaultRegisterer, viper.GetViper())
			return runExport(
				ctx,
				cmd.Root().Version,
				viper.GetViper(),
				prometheus.DefaultRegisterer,
				publisher.SolarEdgeUpdater{SolarEdgeClient: &solarEdgeClient},
				logger,
			)
		},
	}
)

func runExport(
	ctx context.Context,
	version string,
	v *viper.Viper,
	r prometheus.Registerer,
	solarEdgeUpdater publisher.Updater[publisher.SolarEdgeUpdate],
	logger *slog.Logger,
) error {
	logger.Info("starting solaredge exporter", "version", version)
	defer logger.Info("stopping solaredge exporter")

	if pprofAddr := v.GetString("pprof"); pprofAddr != "" {
		go func() {
			logger.Debug("starting pprof", "addr", pprofAddr)
			_ = http.ListenAndServe(pprofAddr, nil)
		}()
	}

	exportMetrics := exporter.NewMetrics()
	r.MustRegister(exportMetrics)

	solarEdgePoller := publisher.Publisher[publisher.SolarEdgeUpdate]{
		Updater:  solarEdgeUpdater,
		Interval: v.GetDuration("polling.interval"),
		Logger:   logger.With("publisher", "solaredge"),
	}

	exp := exporter.Exporter{
		SolarEdge: &solarEdgePoller,
		Metrics:   exportMetrics,
		Logger:    logger,
	}

	var group errgroup.Group
	group.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: v.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	group.Go(func() error { return solarEdgePoller.Run(ctx) })
	group.Go(func() error { return exp.Run(ctx) })

	return group.Wait()
}
