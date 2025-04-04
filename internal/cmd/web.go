package cmd

import (
	"codeberg.org/clambin/go-common/charmer"
	"codeberg.org/clambin/go-common/httputils"
	"codeberg.org/clambin/go-common/httputils/metrics"
	"codeberg.org/clambin/go-common/httputils/middleware"
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
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
	webCmd = cobra.Command{
		Use:   "web",
		Short: "runs the web server",
		PreRun: func(cmd *cobra.Command, args []string) {
			charmer.SetJSONLogger(cmd, viper.GetBool("debug"))
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			logger := charmer.GetLogger(cmd)
			return runWeb(ctx, cmd.Root().Version, viper.GetViper(), prometheus.DefaultRegisterer, logger)
		},
	}
)

func runWeb(ctx context.Context, version string, v *viper.Viper, r prometheus.Registerer, logger *slog.Logger) error {
	logger.Info("starting solaredge web server", "version", version)
	defer logger.Info("stopping solaredge web server")

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

	serverMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "web"})
	r.MustRegister(serverMetrics)

	var cache *web.ImageCache
	if redisClient := newRedisClient(v); redisClient != nil {
		cache = &web.ImageCache{
			Client:    redisClient,
			Namespace: "github.com/clambin/solaredge-monitor",
			Rounding:  v.GetDuration("web.cache.rounding"),
			TTL:       v.GetDuration("web.cache.ttl"),
		}
	}

	h := web.New(repo, cache, logger)
	h = middleware.WithRequestMetrics(serverMetrics)(h)
	h = middleware.RequestLogger(logger, slog.LevelInfo, middleware.DefaultRequestLogFormatter)(h)

	var g errgroup.Group
	g.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: v.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	g.Go(func() error {
		return httputils.RunServer(ctx, &http.Server{Addr: v.GetString("web.addr"), Handler: h})
	})
	return g.Wait()
}
