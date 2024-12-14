package cmd

import (
	"context"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/httputils"
	"github.com/clambin/go-common/httputils/metrics"
	"github.com/clambin/go-common/httputils/middleware"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
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

	serverMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "web"})
	r.MustRegister(serverMetrics)

	var cache *web.ImageCache
	if redisAddr := v.GetString("web.cache.addr"); redisAddr != "" {
		cache = &web.ImageCache{
			Client: redis.NewClient(&redis.Options{
				Addr:     redisAddr,
				Username: v.GetString("web.cache.username"),
				Password: v.GetString("web.cache.password"),
			}),
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
