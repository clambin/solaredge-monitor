package main

import (
	"context"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/monitor"
	"github.com/clambin/solaredge-monitor/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
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

	if scrape {
		cfg.Scrape.Enabled = true
	}

	var m *monitor.Environment
	if m, err = monitor.NewFromConfig(cfg); err != nil {
		log.WithError(err).Fatal("failed to create monitor")
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		m.Run(ctx)
		wg.Done()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs
	cancel()
	wg.Wait()
}
