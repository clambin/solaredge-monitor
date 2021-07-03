package main

import (
	"context"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/poller"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	"github.com/clambin/solaredge-monitor/web/server"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	var (
		debug          bool
		configFileName string
		scrape         bool
		err            error
		cfg            *configuration.Configuration
	)

	a := kingpin.New(filepath.Base(os.Args[0]), "SolarEdge power monitoring")
	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").Short('d').BoolVar(&debug)
	a.Flag("config", "Configuration file").Short('c').Required().StringVar(&configFileName)
	a.Flag("scrape", "Scrape new measurements").Short('s').BoolVar(&scrape)

	if _, err = a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	log.WithField("version", version.BuildVersion).Info("solaredge-monitoring started")

	if cfg, err = configuration.LoadFromFile(configFileName); err != nil {
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

	var (
		coll      *collector.Collector
		power     *poller.SolarEdgePoller
		intensity *poller.TadoPoller
	)

	ctx, cancel := context.WithCancel(context.Background())

	if scrape || cfg.Scrape.Enabled {
		coll = collector.New(cfg.Scrape.Collection, db)
		power = poller.NewSolarEdgePoller(cfg.SolarEdge.Token, coll.Power, cfg.Scrape.Polling)
		intensity = poller.NewTadoPoller(cfg.Tado.Username, cfg.Tado.Password, coll.Intensity, cfg.Scrape.Polling)

		go coll.Run(ctx)
		go power.Run(ctx)
		go intensity.Run(ctx)
	}

	s := server.New(cfg.Server.Port, cfg.Server.Images, reports.New(db))
	go s.Run(ctx)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	cancel()
	if scrape || cfg.Scrape.Enabled {
		// wait for collector to finish last scrape
		time.Sleep(1 * time.Second)
	}
}
