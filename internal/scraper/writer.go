package scraper

import (
	"context"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/tado/v2"
	"log/slog"
	"time"
)

type Writer struct {
	Store
	TadoGetter
	Poller      Publisher[solaredge.Update]
	Interval    time.Duration
	Logger      *slog.Logger
	power       median
	lastWeather *tado.Weather
	HomeId      tado.HomeId
}

type Store interface {
	Store(repository.Measurement) error
}

type TadoGetter interface {
	GetWeatherWithResponse(ctx context.Context, homeId tado.HomeId, reqEditors ...tado.RequestEditorFn) (*tado.GetWeatherResponse, error)
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
			w.process(ctx, update)
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

func (w *Writer) getWeatherInfo(ctx context.Context) error {
	response, err := w.TadoGetter.GetWeatherWithResponse(ctx, w.HomeId)
	if err == nil {
		w.lastWeather = response.JSON200
	}
	return err
}

func (w *Writer) process(ctx context.Context, update solaredge.Update) {
	if err := w.getWeatherInfo(ctx); err != nil {
		w.Logger.Error("failed to get weather info", "err", err)
	}

	if len(update) > 0 {
		w.power.add(update[0].PowerOverview.CurrentPower.Power)
		w.Logger.Debug("update received", "site", update[0].Name, "count", w.power.len())
	}
	if len(update) > 1 {
		w.Logger.Debug("only one site is supported. ignoring remaining sites")
	}

}

func (w *Writer) store() error {
	if w.lastWeather == nil {
		w.Logger.Debug("no weather info to store")
		return nil
	}
	if w.power.len() == 0 {
		w.Logger.Debug("no data to store")
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
		Intensity: float64(*w.lastWeather.SolarIntensity.Percentage),
		Weather:   string(*w.lastWeather.WeatherState.Value),
	}
	w.Logger.Info("storing", "measurement", m)
	return w.Store.Store(m)
}
