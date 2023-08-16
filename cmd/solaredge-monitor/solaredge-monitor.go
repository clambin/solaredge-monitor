package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/go-common/taskmanager/httpserver"
	promserver "github.com/clambin/go-common/taskmanager/prometheus"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/collector/solaredgescraper"
	"github.com/clambin/solaredge-monitor/collector/tadoscraper"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store"
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
		Run:   Main,
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

func Main(_ *cobra.Command, _ []string) {
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

	db, err := store.NewPostgresDB(host, port, database, username, password)
	if err != nil {
		slog.Error("failed to access database", "err", err)
		return
	}
	slog.Info("connected to database", slog.Group("database",
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("database", database),
		slog.String("username", username),
	))

	s := server.New(db)
	prometheus.MustRegister(db, s)

	tasks := []taskmanager.Task{
		promserver.New(promserver.WithAddr(viper.GetString("prometheus.addr"))),
		httpserver.New(viper.GetString("server.addr"), s),
	}

	ctx, done := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer done()

	if viper.GetBool("scrape.enabled") {
		c, err := makeScraper(ctx, db)
		if err != nil {
			slog.Error("failed to create collector", "err", err)
			os.Exit(1)
		}
		tasks = append(tasks, c)
	}

	tm := taskmanager.New(tasks...)

	if err = tm.Run(ctx); err != nil {
		slog.Error("failed to start solaredge-monitor", "err", err)
		os.Exit(1)
	}
}

func makeScraper(ctx context.Context, db store.DB) (*collector.Collector, error) {
	tadoClient, err := tado.NewWithContext(ctx,
		viper.GetString("tado.username"),
		viper.GetString("tado.password"),
		viper.GetString("tado.secret"),
	)
	if err != nil {
		slog.Error("failed to connect to Tado API", "err", err)
		return nil, fmt.Errorf("tado: %w", err)
	}

	site, err := getSite(ctx)
	if err != nil {
		slog.Error("failed to get SolarEdge site", "err", err)
		return nil, fmt.Errorf("solaredge: %w", err)
	}
	c := &collector.Collector{
		TadoScraper:      &tadoscraper.Fetcher{API: tadoClient},
		SolarEdgeScraper: &solaredgescraper.Fetcher{Site: site},
		DB:               db,
		ScrapeInterval:   viper.GetDuration("scrape.polling"),
		CollectInterval:  viper.GetDuration("scrape.collection"),
		Logger:           slog.Default().With("component", "collector"),
	}

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
