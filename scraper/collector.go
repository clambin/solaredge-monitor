package scraper

import (
	"context"
	"github.com/clambin/go-common/taskmanager"
	"github.com/clambin/solaredge-monitor/store"
	"log/slog"
	"time"
)

var _ taskmanager.Task = &Collector{}

type Collector struct {
	Interval       time.Duration
	DB             store.DB
	Logger         *slog.Logger
	WeatherScraper Publisher[WeatherInfo]
	solarIntensity Averager
	temperature    Averager
	weather        WordCounter
	PowerScraper   Publisher[PowerInfo]
	power          Averager
}

func (c *Collector) Run(ctx context.Context) error {
	weatherCh := c.WeatherScraper.Subscribe()
	defer c.WeatherScraper.Unsubscribe(weatherCh)

	powerCh := c.PowerScraper.Subscribe()
	defer c.PowerScraper.Unsubscribe(powerCh)

	ticker := time.NewTicker(c.Interval)
	defer ticker.Stop()

	c.Logger.Debug("starting")
	for {
		select {
		case info := <-weatherCh:
			c.solarIntensity.Add(info.SolarIntensity)
			c.temperature.Add(info.Temperature)
			c.weather.Add(info.Weather)
		case info := <-powerCh:
			c.power.Add(info.Power)
		case <-ticker.C:
			c.collect()
		case <-ctx.Done():
			c.collect()
			c.Logger.Debug("stopping")
			return ctx.Err()
		}
	}
}

func (c *Collector) collect() {
	if c.power.Count == 0 || c.solarIntensity.Count == 0 {
		c.Logger.Warn("partial data collection. skipping")
		return
	}

	collectorSamples.WithLabelValues("power").Set(float64(c.power.Count))
	collectorSamples.WithLabelValues("solar").Set(float64(c.solarIntensity.Count))

	measurement := store.Measurement{
		Timestamp: time.Now(),
		Power:     c.power.Average(),
		Intensity: c.solarIntensity.Average(),
		Weather:   c.weather.GetMostUsed(),
	}

	if measurement.Power == 0 && measurement.Intensity == 0 {
		c.Logger.Debug("no solar power activity. skipping measurement")
		return
	}

	if err := c.DB.Store(measurement); err != nil {
		c.Logger.Error("failed to store metrics", "err", err)
		return
	}

	c.Logger.Info("new entry", slog.Group("measurement",
		slog.Float64("power", measurement.Power),
		slog.Float64("intensity", measurement.Intensity),
		slog.String("weather", measurement.Weather),
	))
}
