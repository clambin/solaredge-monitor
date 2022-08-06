package monitor_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/clambin/solaredge-monitor/monitor"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNewFromConfigWithDB(t *testing.T) {
	db := mockdb.BuildDB()
	config := configuration.Configuration{
		Server: configuration.ServerConfiguration{
			Port: 8080,
		},
		Scrape: configuration.ScrapeConfiguration{
			Enabled:    true,
			Polling:    10 * time.Minute,
			Collection: time.Hour,
		},
		Tado: configuration.TadoConfiguration{
			Username: "foo",
			Password: "bar",
		},
		SolarEdge: configuration.SolarEdgeConfiguration{
			Token: "foo",
		},
	}

	m, err := monitor.NewFromConfigWithDB(&config, db)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.NotNil(t, m.Server)

	config.Scrape.Enabled = false
	m, err = monitor.NewFromConfigWithDB(&config, db)
	require.NoError(t, err)
	require.NotNil(t, m)
	assert.Nil(t, m.Collector)
}

func TestNewFromConfig(t *testing.T) {
	config := configuration.Configuration{Database: configuration.DBConfiguration{
		Host:     "localhost",
		Port:     5432,
		Database: "foo",
		Username: "foo",
		Password: "bar",
	}}
	_, err := monitor.NewFromConfig(&config)
	assert.Error(t, err)
	dbError := driver.ErrBadConn
	assert.True(t, errors.As(err, &dbError))
}

type fakeCollector struct {
	Called bool
}

func (f *fakeCollector) Run(ctx context.Context) {
	f.Called = true
	<-ctx.Done()
}

func TestEnvironment_Run(t *testing.T) {
	db := mockdb.BuildDB()
	c := &fakeCollector{}
	e := &monitor.Environment{
		DB:        db,
		Server:    server.New(8080, db),
		Collector: c,
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		e.Run(ctx)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		_, err := http.Get("http://localhost:8080/")
		return err == nil
	}, 5*time.Second, 100*time.Millisecond)

	cancel()
	wg.Wait()

	assert.True(t, c.Called)
}
