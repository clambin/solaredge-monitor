package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/charmer"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"time"
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

	go func() {
		err := http.ListenAndServe(viper.GetString("prometheus.addr"), promhttp.Handler())
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("failed to start prometheus server", "err", err)
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

	serverMetrics := metrics.NewRequestSummaryMetrics("solaredge", "web", nil)
	prometheus.MustRegister(serverMetrics)

	h := web.New(repo, logger)
	h = middleware.WithRequestMetrics(serverMetrics)(h)
	h = middleware.RequestLogger(logger.With("component", "web"), slog.LevelInfo, middleware.DefaultRequestLogFormatter)(h)

	server := http.Server{Addr: viper.GetString("web.addr"), Handler: h}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("error starting web server", "err", err)
		}
	}()

	<-cmd.Context().Done()

	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		logger.Error("error shutting down web server", "err", err)
	}

	return err
}
