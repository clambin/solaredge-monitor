package scraper_test

/*
import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/solaredge-monitor/scraper"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
	"time"
)

func TestCollector(t *testing.T) {
	db := mockdb.NewDB()

	powerScraper := scraper.PowerScraper{
		PowerGetter: newMockPowerGetter(t),
		Logger:      slog.Default().With("component", "powerScraper"),
		Interval:    50 * time.Millisecond,
	}

	weatherScraper := scraper.WeatherScraper{
		WeatherGetter: newMockWeatherGetter(t),
		Logger:        slog.Default().With("component", "weatherScraper"),
		Interval:      50 * time.Millisecond,
	}

	c := scraper.Collector{
		PowerScraper:   &powerScraper,
		WeatherScraper: &weatherScraper,
		DB:             db,
		Interval:       100 * time.Millisecond,
		Logger:         slog.Default().With("component", "collector"),
	}

	tasks := taskmanager.New(&c, &weatherScraper, &powerScraper)

	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan error)
	go func() { ch <- tasks.Run(ctx) }()

	assert.Eventually(t, func() bool { return db.Rows() > 0 }, time.Second, 100*time.Millisecond)

	measurements, _ := db.GetAll()
	for _, entry := range measurements {
		assert.Equal(t, 1000.0, entry.Power)
		assert.Equal(t, 75.0, entry.Intensity)
		assert.Equal(t, "SUNNY", entry.Weather)
	}

	cancel()
	assert.ErrorIs(t, <-ch, context.Canceled)
}


*/
