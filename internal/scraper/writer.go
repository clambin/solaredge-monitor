package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/publisher/solaredge"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/tado/v2"
	"log/slog"
	"time"
)

type Writer struct {
	Store
	SolarEdge      Publisher[solaredge.Update]
	Tado           Publisher[*tado.Weather]
	Interval       time.Duration
	Logger         *slog.Logger
	power          median
	solarIntensity median
	weatherStates  weatherStates
}

type Publisher[T any] interface {
	Subscribe() <-chan T
	Unsubscribe(<-chan T)
}

type Store interface {
	Store(repository.Measurement) error
}

func (w *Writer) Run(ctx context.Context) error {
	w.Logger.Debug("starting writer", "interval", w.Interval)
	defer w.Logger.Debug("stopped writer")

	solarEdgeUpdate := w.SolarEdge.Subscribe()
	defer w.SolarEdge.Unsubscribe(solarEdgeUpdate)

	tadoUpdate := w.Tado.Subscribe()
	defer w.Tado.Unsubscribe(tadoUpdate)

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case update := <-solarEdgeUpdate:
			w.processSolarEdgeUpdate(update)
		case update := <-tadoUpdate:
			w.processTadoUpdate(update)
		case <-ticker.C:
			if err := w.store(); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
		case <-ctx.Done():
			w.Logger.Debug("shutting down. saving partial data")
			if err := w.store(); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
			return nil
		}
	}
}

func (w *Writer) processSolarEdgeUpdate(update solaredge.Update) {
	if len(update) > 0 {
		w.power.add(update[0].PowerOverview.CurrentPower.Power)
		w.Logger.Debug("update received", "site", update[0].Name, "count", w.power.len())
	}
	if len(update) > 1 {
		w.Logger.Debug("only one site is supported. ignoring remaining sites")
	}
}

func (w *Writer) processTadoUpdate(update *tado.Weather) {
	w.solarIntensity.add(float64(*update.SolarIntensity.Percentage))
	w.weatherStates = append(w.weatherStates, string(*update.WeatherState.Value))
}

func (w *Writer) store() error {
	if w.solarIntensity.len() == 0 {
		w.Logger.Debug("no weather info to store")
		return nil
	}
	if w.power.len() == 0 {
		w.Logger.Debug("no power data to store")
		return nil
	}
	power := w.power.median()
	if power == 0 {
		w.Logger.Debug("not storing measurement with no power")
		return nil
	}

	m := repository.Measurement{
		Timestamp: time.Now(),
		Power:     power,
		Intensity: w.solarIntensity.median(),
		Weather:   w.weatherStates.mostFrequent(),
	}

	w.power.reset()
	w.solarIntensity.reset()
	w.weatherStates = w.weatherStates[:0]

	w.Logger.Info("storing", "measurement", m)
	return w.Store.Store(m)
}
