package collector_test

import (
	"context"
	"github.com/clambin/solaredge"
	"github.com/clambin/solaredge-monitor/collector"
	solaredgeScraper "github.com/clambin/solaredge-monitor/collector/solaredgeScraper"
	"github.com/clambin/solaredge-monitor/collector/solaredgeScraper/mocks"
	tadoScraper "github.com/clambin/solaredge-monitor/collector/tadoScraper"
	tadoMock "github.com/clambin/solaredge-monitor/collector/tadoScraper/mocks"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/clambin/tado"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"log/slog"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	db := mockdb.NewDB()
	tadoClient := tadoMock.NewAPI(t)
	site := mocks.NewSite(t)

	c := collector.Collector{
		TadoScraper:      &tadoScraper.Fetcher{API: tadoClient},
		SolarEdgeScraper: &solaredgeScraper.Fetcher{Site: site},
		DB:               db,
		ScrapeInterval:   10 * time.Millisecond,
		CollectInterval:  100 * time.Millisecond,
		Logger:           slog.Default(),
	}

	tadoClient.EXPECT().GetWeatherInfo(mock.Anything).Return(tado.WeatherInfo{
		SolarIntensity: tado.Percentage{Percentage: 55.0},
		WeatherState:   tado.Value{Value: "SUNNY"},
	}, nil)
	site.EXPECT().GetPowerOverview(mock.Anything).Return(solaredge.PowerOverview{
		LastUpdateTime: solaredge.Time{},
		LifeTimeData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 10000,
		},
		LastYearData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 1000,
		},
		LastMonthData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 100,
		},
		LastDayData: struct {
			Energy  float64 `json:"energy"`
			Revenue float64 `json:"revenue"`
		}{
			Energy: 10,
		},
		CurrentPower: struct {
			Power float64 `json:"power"`
		}{
			Power: 3400,
		},
	}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- c.Run(ctx) }()

	assert.Eventually(t, func() bool { return db.Rows() > 0 }, time.Second, 50*time.Millisecond)

	measurements, _ := db.GetAll()
	for _, entry := range measurements {
		assert.Equal(t, 3400.0, entry.Power)
		assert.Equal(t, 55.0, entry.Intensity)
	}

	cancel()
	assert.NoError(t, <-ch)
}
