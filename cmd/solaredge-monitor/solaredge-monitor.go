package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/go-common/taskmanager/httpserver"
	promserver "github.com/clambin/go-common/taskmanager/prometheus"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/internal/analyzer"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper"
	server "github.com/clambin/solaredge-monitor/internal/web"
	"github.com/clambin/tado"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	version = "change_me"

	configFile string
	cmd        = cobra.Command{
		Use:   "solaredge-monitor",
		Short: "records solar panel output vs. weather conditions",
		RunE:  Main,
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to start", "err", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cmd.Version = version
	cmd.Flags().StringVar(&configFile, "config", "", "Configuration file")
	cmd.Flags().Bool("debug", false, "Log debug messages")
	cmd.Flags().Bool("scrape", false, "Record measurements")
	_ = viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	_ = viper.BindPFlag("scrape.enabled", cmd.Flags().Lookup("scrape"))
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("/etc/solaredge/")
		viper.AddConfigPath("$HOME/.solaredge")
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("prometheus.addr", ":9090")
	viper.SetDefault("scrape.polling", 5*time.Minute)
	viper.SetDefault("scrape.collection", 15*time.Minute)
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.database", "solar")
	viper.SetDefault("database.username", "solar")
	viper.SetDefault("database.password", "")
	viper.SetDefault("solaredge.token", "")
	viper.SetDefault("tado.username", "")
	viper.SetDefault("tado.password", "")
	viper.SetDefault("tado.secret", "")

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config file", "err", err)
		os.Exit(1)
	}
}

func Main(_ *cobra.Command, _ []string) error {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &opts)))

	slog.Info("solaredge-monitor started", "version", version)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := repository.NewPostgresDB(host, port, database, username, password)
	if err != nil {
		return fmt.Errorf("failed to access database: %w", err)
	}

	slog.Debug("connected to database", slog.Group("database",
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("database", database),
		slog.String("username", username),
	))

	s := server.NewHTTPServer(db, slog.With("component", "webserver"))
	prometheus.MustRegister(db, s.PrometheusMetrics)

	tasks := []taskmanager.Task{
		promserver.New(promserver.WithAddr(viper.GetString("prometheus.addr"))),
		httpserver.New(viper.GetString("server.addr"), s.Router),
		&analyzer.WeatherDaemon{
			Repository: db,
			Interval:   time.Hour,
			Logger:     slog.Default().With("component", "analyzer"),
		},
	}

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	if viper.GetBool("scrape.enabled") {
		c, err := makeScraper(ctx, db)
		if err != nil {
			return fmt.Errorf("failed to create scraper: %w", err)
		}
		tasks = append(tasks, c)
	}

	tm := taskmanager.New(tasks...)

	if err = tm.Run(ctx); errors.Is(err, context.Canceled) {
		err = nil
	}
	return err
}

func makeScraper(ctx context.Context, db *repository.PostgresDB) (*taskmanager.Manager, error) {
	tadoClient, err := tado.NewWithContext(ctx,
		viper.GetString("tado.username"),
		viper.GetString("tado.password"),
		viper.GetString("tado.secret"),
	)
	if err != nil {
		return nil, fmt.Errorf("tado: %w", err)
	}

	site, err := getSite(ctx)
	if err != nil {
		return nil, fmt.Errorf("solaredge: %w", err)
	}
	c := scraper.New(
		site, tadoClient, db,
		slog.Default().With("component", "scraper"),
		viper.GetDuration("scrape.polling"),
		viper.GetDuration("scrape.collection"),
	)

	return c, nil
}

func getSite(ctx context.Context) (*solaredge.Site, error) {
	c := solaredge.Client{
		Token:      viper.GetString("solaredge.token"),
		HTTPClient: http.DefaultClient,
	}
	sites, err := c.GetSites(ctx)
	if err != nil {
		return nil, err
	}
	if len(sites) == 0 {
		return nil, errors.New("no SolarEdge sites found")
	}
	if len(sites) > 1 {
		slog.Warn("more than one SolarEdge site found. Using first one", "name", sites[0].Name)
	}
	return &sites[0], nil
}
