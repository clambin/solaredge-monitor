package scraper

import (
	"context"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/scraper/solaredge"
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
	power      Averager
}

type Store interface {
	Store(repository.Measurement) error
}

type TadoGetter interface {
	GetWeatherInfo(context.Context) (tado.WeatherInfo, error)
}

func (w *Writer) Run(ctx context.Context) error {
	ch := w.Poller.Subscribe()
	defer w.Poller.Unsubscribe(ch)

	for {
		select {
		case update := <-ch:
			w.process(update)
		case <-time.After(w.Interval):
			if err := w.save(ctx); err != nil {
				w.Logger.Error("failed to save update", "err", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (w *Writer) process(update solaredge.Update) {
	for site := range update {
		if site > 0 {
			return
		}
		w.power.Add(update[0].PowerOverview.CurrentPower.Power)
	}
}

func (w *Writer) save(ctx context.Context) error {
	if w.power.Count == 0 {
		return fmt.Errorf("no measurements available")
	}

	c, err := w.TadoClient.GetWeatherInfo(ctx)
	if err != nil {
		return fmt.Errorf("tado: %w", err)
	}

	m := repository.Measurement{
		Timestamp: time.Now(),
		Power:     w.power.Average(),
		Intensity: c.SolarIntensity.Percentage,
		Weather:   c.WeatherState.Value,
	}
	return w.Store.Store(m)
}
