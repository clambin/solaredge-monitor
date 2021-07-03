package main

import (
	"context"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/feeder"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/solaredge-monitor/version"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var (
		debug          bool
		configFileName string
		err            error
		cfg            *configuration.Configuration
		prometheusURL  string
	)

	a := kingpin.New(filepath.Base(os.Args[0]), "SolarEdge power monitoring backfiller")
	a.Version(version.BuildVersion)
	a.HelpFlag.Short('h')
	a.VersionFlag.Short('v')
	a.Flag("debug", "Log debug messages").Short('d').BoolVar(&debug)
	a.Flag("config", "Configuration file").Short('c').Required().StringVar(&configFileName)
	a.Flag("prometheus", "Prometheus URL").Short('p').Required().StringVar(&prometheusURL)

	if _, err = a.Parse(os.Args[1:]); err != nil {
		a.Usage(os.Args[1:])
		os.Exit(1)
	}

	log.WithField("version", version.BuildVersion).Info("backfiller started")

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
		power     []collector.Metric
		intensity []collector.Metric
	)

	stop := time.Now()
	start := stop.Add(-7 * 24 * time.Hour)

	if power, err = feeder.GetPowerMetrics(prometheusURL, start, stop); err != nil {
		log.WithError(err).Error("failed to load power data")
	}

	if intensity, err = feeder.GetIntensityMetrics(prometheusURL, start, stop); err != nil {
		log.WithError(err).Error("failed to load solar intensity data")
	}

	log.Infof("discovered %d power metrics", len(power))
	log.Infof("discovered %d solar intensity metrics", len(intensity))

	coll := collector.New(cfg.Scrape.Collection, db)
	ctx, cancel := context.WithCancel(context.Background())
	go coll.Run(ctx)

	if err = feeder.FeedMetrics(power, intensity, coll); err != nil {
		log.WithError(err).Error("failed to feed metrics")
	}

	cancel()
	time.Sleep(1)
}
