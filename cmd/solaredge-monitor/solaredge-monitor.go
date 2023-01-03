package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	"github.com/clambin/tado"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/exp/slog"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
)

func main() {
	var (
		debug          bool
		configFileName string
		scrape         bool
	)

	a := kingpin.New(filepath.Base(os.Args[0]), "SolarEdge power monitoring")
	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").Short('d').BoolVar(&debug)
	a.Flag("config", "Configuration file").Short('c').Required().StringVar(&configFileName)
	a.Flag("scrape", "Scrape new measurements").Short('s').BoolVar(&scrape)

	if _, err := a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	cfg, err := configuration.LoadFromFile(configFileName)
	if err != nil {
		slog.Error("failed to read configuration file", err)
		panic(err)
	}

	var opts slog.HandlerOptions
	if debug || cfg.Debug {
		opts.Level = slog.LevelDebug
		opts.AddSource = true
	}
	slog.SetDefault(slog.New(opts.NewTextHandler(os.Stdout)))

	slog.Info("solaredge-monitoring started", "version", version.BuildVersion)

	if scrape {
		cfg.Scrape.Enabled = true
	}

	db, err := store.NewPostgresDBFromConfig(cfg.Database)
	if err != nil {
		slog.Error("failed to access database", err)
		panic(err)
	}
	s := server.New(cfg.Server.Port, db)

	prometheus.MustRegister(db, s)

	ctx, cancel := context.WithCancel(context.Background())
	go s.Run(ctx)

	var wg sync.WaitGroup
	if cfg.Scrape.Enabled {
		wg.Add(1)
		go func() {
			runScraper(ctx, cfg, db)
			wg.Done()
		}()
	}

	go runPrometheusServer(cfg.Server.PrometheusPort)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
	wg.Wait()
}

func runPrometheusServer(port int) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); !errors.Is(err, http.ErrServerClosed) {
		slog.Error("could not start prometheus metrics server", err)
		panic(err)
	}
}

func runScraper(ctx context.Context, cfg *configuration.Configuration, db store.DB) {
	c := &collector.Collector{
		Tado: &scraper.Client{
			Scraper: &scraper.TadoScraper{
				API: tado.New(cfg.Tado.Username, cfg.Tado.Password, ""),
			},
			Interval: cfg.Scrape.Polling,
		},
		SolarEdge: &scraper.Client{
			Scraper: &scraper.SolarEdgeScraper{
				API: &solaredge.Client{
					Token:      cfg.SolarEdge.Token,
					HTTPClient: http.DefaultClient,
				},
			},
			Interval: cfg.Scrape.Polling,
		},
		DB:       db,
		Interval: cfg.Scrape.Collection,
	}
	c.Run(ctx)
}
