package web

import (
	"errors"
	"fmt"
	"github.com/clambin/go-common/http/metrics"
	"github.com/clambin/go-common/http/middleware"
	"github.com/clambin/solaredge-monitor/internal/repository"
	server "github.com/clambin/solaredge-monitor/internal/web"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
)

var (
	Cmd = cobra.Command{
		Use:   "web",
		Short: "runs the web server",
		RunE:  run,
	}
)

func run(cmd *cobra.Command, _ []string) error {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &opts))

	go func() {
		err := http.ListenAndServe(viper.GetString("prometheus.addr"), promhttp.Handler())
		if !errors.Is(err, http.ErrServerClosed) {
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

	mw1 := middleware.RequestLogger(logger.With("component", "web"), slog.LevelInfo, middleware.DefaultRequestLogFormatter)
	mw2 := middleware.WithRequestMetrics(serverMetrics)

	logger.Info("starting solaredge web server", "version", cmd.Root().Version)
	defer logger.Info("stopping solaredge web server")

	err = http.ListenAndServe(viper.GetString("web.addr"), mw1(mw2(server.New(repo, logger))))
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}
