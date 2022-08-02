package main

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	"github.com/clambin/solaredge-monitor/web/server"
	"github.com/clambin/tado"
	log "github.com/sirupsen/logrus"
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

	log.WithField("version", version.BuildVersion).Info("solaredge-monitoring started")

	cfg, err := configuration.LoadFromFile(configFileName)
	if err != nil {
		log.WithError(err).Fatal("failed to read configuration file")
	}

	if debug || cfg.Debug {
		log.SetLevel(log.DebugLevel)
	}

	db := store.NewPostgresDB(
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Database,
		cfg.Database.Username,
		cfg.Database.Password,
	)

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}

	if scrape || cfg.Scrape.Enabled {
		wg.Add(1)
		go func() {
			runScraper(ctx, cfg, db)
			wg.Done()
		}()
	}

	s := server.New(cfg.Server.Port, cfg.Server.Images, reports.New(db))
	go s.Run(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
	wg.Wait()
}

func runScraper(ctx context.Context, cfg *configuration.Configuration, db store.DB) {
	tadoClient := &scraper.Client{
		Scraper: &scraper.TadoScraper{
			API: tado.New(cfg.Tado.Username, cfg.Tado.Password, ""),
		},
	}
	go tadoClient.Run(ctx, cfg.Scrape.Polling)

	solarEdgeClient := &scraper.Client{
		Scraper: &scraper.SolarEdgeScraper{
			API: &solaredge.Client{
				Token:      cfg.SolarEdge.Token,
				HTTPClient: http.DefaultClient,
			},
		},
	}
	go solarEdgeClient.Run(ctx, cfg.Scrape.Polling)

	coll := collector.Collector{
		SolarEdge: solarEdgeClient,
		Tado:      tadoClient,
		DB:        db,
	}
	coll.Run(ctx, cfg.Scrape.Collection)
}
