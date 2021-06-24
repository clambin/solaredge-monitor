package main

import (
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/poller"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

func main() {
	var (
		debug          bool
		configFileName string
		err            error
		cfg            *configuration.Configuration
	)
	log.WithField("version", version.BuildVersion).Info("solaredge-monitoring started")

	a := kingpin.New(filepath.Base(os.Args[0]), "SolarEdge power monitoring")
	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").Short('d').BoolVar(&debug)
	a.Flag("config", "Configuration file").Short('c').Required().StringVar(&configFileName)

	if _, err = a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

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

	coll := collector.New(cfg.Collection, db)
	power := poller.NewSolarEdgePoller(cfg.SolarEdge.Token, coll.Power, cfg.Polling)
	intensity := poller.NewTadoPoller(cfg.Tado.Username, cfg.Tado.Password, coll.Intensity, cfg.Polling)

	go power.Run()
	go intensity.Run()
	coll.Run()
}
