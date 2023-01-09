package collector_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/collector"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	mockSolaredge "github.com/clambin/solaredge/mocks"
	"github.com/clambin/tado"
	mockTado "github.com/clambin/tado/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	db := mockdb.NewDB()
	solarEdgeClient := mockSolaredge.NewAPI(t)
	tadoClient := mockTado.NewAPI(t)

	c := collector.Collector{
		TadoAPI:      tadoClient,
		SolarEdgeAPI: solarEdgeClient,
		DB:           db,
	}

	solarEdgeClient.
		On("GetSiteIDs", mock.AnythingOfType("*context.cancelCtx")).
		Return([]int{100}, nil)
	solarEdgeClient.
		On("GetPowerOverview", mock.AnythingOfType("*context.cancelCtx"), 100).
		Return(0.0, 0.0, 0.0, 0.0, 1500.0, nil)
	tadoClient.
		On("GetWeatherInfo", mock.AnythingOfType("*context.cancelCtx")).
		Return(tado.WeatherInfo{
			SolarIntensity: tado.Percentage{Percentage: 55.0},
			WeatherState:   tado.Value{Value: "SUNNY"},
		}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { defer wg.Done(); c.Run(ctx, 10*time.Millisecond, 100*time.Millisecond) }()

	assert.Eventually(t, func() bool { return db.Rows() > 0 }, 500*time.Millisecond, 50*time.Millisecond)

	measurements, _ := db.GetAll()
	for _, entry := range measurements {
		assert.Equal(t, 1500.0, entry.Power)
		assert.Equal(t, 55.0, entry.Intensity)
	}

	cancel()
	wg.Wait()
}