package web

import (
	"fmt"
	"github.com/clambin/go-common/charmer"
	gchttp "github.com/clambin/go-common/http"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
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
	Cmd = cobra.Command{
		Use:   "web",
		Short: "runs the web server",
		RunE:  run,
	}
)

func run(cmd *cobra.Command, _ []string) error {
	logger := charmer.GetLogger(cmd)

	logger.Info("starting solaredge web server", "version", cmd.Root().Version)
	defer logger.Info("stopping solaredge web server")

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

	serverMetrics := metrics.NewRequestMetrics(metrics.Options{Namespace: "solaredge", Subsystem: "web"})
	prometheus.MustRegister(serverMetrics)

	var cache *web.ImageCache
	if redisAddr := viper.GetString("web.cache.addr"); redisAddr != "" {
		cache = &web.ImageCache{
			Client: redis.NewClient(&redis.Options{
				Addr:     redisAddr,
				Username: viper.GetString("web.cache.username"),
				Password: viper.GetString("web.cache.password"),
			}),
			Namespace: "github.com/clambin/solaredge-monitor",
			Rounding:  viper.GetDuration("web.cache.rounding"),
			TTL:       viper.GetDuration("web.cache.ttl"),
		}
	}

	h := web.New(repo, cache, logger)
	h = middleware.WithRequestMetrics(serverMetrics)(h)
	h = middleware.RequestLogger(logger.With("component", "web"), slog.LevelInfo, middleware.DefaultRequestLogFormatter)(h)

	var g errgroup.Group
	g.Go(func() error {
		return gchttp.RunServer(cmd.Context(), &http.Server{Addr: viper.GetString("prometheus.addr"), Handler: promhttp.Handler()})
	})
	g.Go(func() error {
		return gchttp.RunServer(cmd.Context(), &http.Server{Addr: viper.GetString("web.addr"), Handler: h})
	})
	return g.Wait()
}
