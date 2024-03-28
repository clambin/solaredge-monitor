package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/averager"
	"github.com/clambin/tado"
	"log/slog"
	"time"
)

type Writer struct {
	Store
	TadoGetter
	Poller      Publisher[solaredge.Update]
	Interval    time.Duration
	Logger      *slog.Logger
	power       averager.Averager[float64]
	lastWeather *tado.WeatherInfo
}

type Store interface {
	Store(repository.Measurement) error
}

type TadoGetter interface {
	GetWeatherInfo(context.Context) (tado.WeatherInfo, error)
}

func (w *Writer) Run(ctx context.Context) error {
	w.Logger.Debug("starting writer", "interval", w.Interval)
	defer w.Logger.Debug("stopped writer")

	ch := w.Poller.Subscribe()
	defer w.Poller.Unsubscribe(ch)

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case update := <-ch:
			w.process(update)
		case <-ticker.C:
			if err := w.getWeatherInfo(ctx); err != nil {
				w.Logger.Error("failed to get weather info", "err", err)
				break
			}
			if err := w.store(); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
		case <-ctx.Done():
			if err := w.store(); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
			return nil
		}
	}
}

func (w *Writer) getWeatherInfo(ctx context.Context) error {
	weatherInfo, err := w.TadoGetter.GetWeatherInfo(ctx)
	if err == nil {
		w.lastWeather = &weatherInfo
	}
	return err
}

func (w *Writer) process(update solaredge.Update) {
	for site := range update {
		if site > 0 {
			w.Logger.Debug("only one site is supported. ignoring remaining sites")
			return
		}
		w.power.Add(update[0].PowerOverview.CurrentPower.Power)
		w.Logger.Debug("update received", "site", update[site].Name, "count", w.power.Count())
	}
}

func (w *Writer) store() error {
	if w.lastWeather == nil || w.power.Count() == 0 {
		w.Logger.Debug("no data to store")
		return nil
	}

	power := w.power.Average()
	if power == 0 {
		w.Logger.Debug("not storing measurement with no power")
		return nil
	}

	m := repository.Measurement{
		Timestamp: time.Now(),
		Power:     power,
		Intensity: w.lastWeather.SolarIntensity.Percentage,
		Weather:   w.lastWeather.WeatherState.Value,
	}
	w.Logger.Info("storing", "measurement", m)
	return w.Store.Store(m)
}
