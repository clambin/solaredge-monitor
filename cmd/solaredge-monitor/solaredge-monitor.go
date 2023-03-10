package main

import (
	"context"
	"errors"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/collector/solaredgescraper"
	"github.com/clambin/solaredge-monitor/collector/tadoscraper"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	"github.com/clambin/tado"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	configFile string
	cmd        = cobra.Command{
		Use:   "solaredge-monitor",
		Short: "records solar panel output vs. weather conditions",
		Run:   Main,
	}
)

func main() {
	if err := cmd.Execute(); err != nil {
		slog.Error("failed to start", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	cmd.Version = version.BuildVersion
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

	viper.SetEnvPrefix("SOLAREDGE")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		slog.Error("failed to read config file", err)
		os.Exit(1)
	}
}

func Main(_ *cobra.Command, _ []string) {
	var opts slog.HandlerOptions
	if viper.GetBool("debug") {
		opts.Level = slog.LevelDebug
		//opts.AddSource = true
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stdout)))

	slog.Info("solaredge-monitor started", "version", version.BuildVersion)

	host := viper.GetString("database.host")
	port := viper.GetInt("database.port")
	database := viper.GetString("database.database")
	username := viper.GetString("database.username")
	password := viper.GetString("database.password")

	db, err := store.NewPostgresDB(host, port, database, username, password)
	if err != nil {
		slog.Error("failed to access database", err)
		return
	}
	slog.Info("connected to database", slog.Group("database",
		slog.String("host", host),
		slog.Int("port", port),
		slog.String("database", database),
		slog.String("username", username),
	))

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	if viper.GetBool("scrape.enabled") {
		wg.Add(1)
		go func() {
			runScraper(ctx, db)
			wg.Done()
		}()
	}

	s := server.New(db)
	prometheus.MustRegister(db, s)

	go runServer(s)
	go runPrometheusServer()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
	wg.Wait()
}

func runServer(s *server.Server) {
	addr := viper.GetString("server.addr")
	slog.Info("starting server", "addr", addr)
	if err := http.ListenAndServe(addr, s); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("could not start server", err)
	}
}

func runPrometheusServer() {
	addr := viper.GetString("prometheus.addr")
	slog.Info("starting Prometheus metrics server", "addr", addr)
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(addr, nil); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("could not start Prometheus metrics server", err)
	}
}

func runScraper(ctx context.Context, db store.DB) {
	site, err := getSite(ctx)
	if err != nil {
		slog.Error("failed to get SolarEdge site", err)
		return
	}
	c := &collector.Collector{
		TadoScraper: &tadoscraper.Fetcher{API: tado.New(
			viper.GetString("tado.username"),
			viper.GetString("tado.password"),
			"",
		)},
		SolarEdgeScraper: &solaredgescraper.Fetcher{Site: site},
		DB:               db,
	}

	polling := viper.GetDuration("scrape.polling")
	collect := viper.GetDuration("scrape.collection")

	slog.Info("starting scraper", slog.Group("scrape",
		slog.Duration("poll", polling),
		slog.Duration("collect", collect),
	))
	c.Run(ctx, polling, collect)
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
