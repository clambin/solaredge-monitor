package monitor

import (
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/scrape/collector"
	"github.com/clambin/solaredge-monitor/scrape/scraper"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/clambin/tado"
	"golang.org/x/net/context"
	"net/http"
	"sync"
)

type Environment struct {
	DB        store.DB
	Server    *server.Server
	Collector *collector.Collector
}

func NewFromConfig(config *configuration.Configuration) (e *Environment, err error) {
	var db store.DB
	if db, err = newDB(config); err != nil {
		return
	}

	e = &Environment{
		DB:     db,
		Server: server.New(config.Server.Port, db),
	}

	if config.Scrape.Enabled {
		e.Collector = &collector.Collector{
			SolarEdge: &scraper.Client{
				Scraper: &scraper.TadoScraper{
					API: tado.New(config.Tado.Username, config.Tado.Password, ""),
				},
				Interval: config.Scrape.Polling,
			},
			Tado: &scraper.Client{
				Scraper: &scraper.SolarEdgeScraper{
					API: &solaredge.Client{
						Token:      config.SolarEdge.Token,
						HTTPClient: http.DefaultClient,
					},
				},
				Interval: config.Scrape.Polling,
			},
			DB:       db,
			Interval: config.Scrape.Collection,
		}

	}
	return
}

func newDB(config *configuration.Configuration) (store.DB, error) {
	return store.NewPostgresDB(
		config.Database.Host, config.Database.Port,
		config.Database.Database,
		config.Database.Username, config.Database.Password,
	)
}

func (e *Environment) Run(ctx context.Context) {
	wg := sync.WaitGroup{}
	if e.Collector != nil {
		wg.Add(1)
		go func() {
			e.Collector.Run(ctx)
			wg.Done()
		}()
	}

	e.Server.Run(ctx)
	wg.Wait()
}
