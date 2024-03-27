package scraper

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
	"github.com/clambin/solaredge-monitor/pkg/averager"
	"github.com/clambin/tado"
	"log/slog"
	"time"
)

type Writer struct {
	Store      Store
	TadoClient TadoGetter
	Poller     Publisher[solaredge.Update]
	Interval   time.Duration
	Logger     *slog.Logger
	power      averager.Averager[float64]
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
			if err := w.store(ctx); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
		case <-ctx.Done():
			if err := w.store(ctx); err != nil {
				w.Logger.Error("failed to store update", "err", err)
			}
			return nil
		}
	}
}

func (w *Writer) process(update solaredge.Update) {
	for site := range update {
		if site > 0 {
			w.Logger.Debug("only one site is supporter. ignoring remaining sites")
			return
		}
		w.power.Add(update[0].PowerOverview.CurrentPower.Power)
		w.Logger.Debug("update received", "site", update[site].Name, "count", w.power.Count)
	}
}

func (w *Writer) store(ctx context.Context) error {
	if w.power.Count() == 0 {
		w.Logger.Debug("no data to store")
		return nil
	}

	power := w.power.Average()
	if power == 0 {
		w.Logger.Debug("not storing measurement with no power")
		return nil
	}

	c, err := w.TadoClient.GetWeatherInfo(ctx)
	if err != nil {
		return fmt.Errorf("tado: %w", err)
	}

	m := repository.Measurement{
		Timestamp: time.Now(),
		Power:     power,
		Intensity: c.SolarIntensity.Percentage,
		Weather:   c.WeatherState.Value,
	}
	w.Logger.Debug("storing", "measurement", m)
	return w.Store.Store(m)
}
