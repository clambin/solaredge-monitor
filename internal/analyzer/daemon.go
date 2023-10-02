package analyzer

import (
	"context"
	"errors"
	"fmt"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/evaluation"
	"log/slog"
	"strconv"
	"time"
)

type Getter interface {
	Get(time.Time, time.Time) (repository.Measurements, error)
}

type WeatherDaemon struct {
	Repository Getter
	Interval   time.Duration
	Logger     *slog.Logger
}

func (w WeatherDaemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	w.Logger.Debug("started")
	defer w.Logger.Debug("stopped")

	for {
		select {
		case <-ticker.C:
			if err := w.learn(); err != nil {
				w.Logger.Error("failed to learn weather pattern", "err", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (w WeatherDaemon) learn() error {
	measurements, err := w.Repository.Get(time.Time{}, time.Time{})
	if err != nil {
		return fmt.Errorf("db: %w", err)
	}

	measurements = removeUnknown(measurements)

	allData := AnalyzeMeasurements(measurements)
	trainData, testData := base.InstancesTrainTestSplit(allData, .95)
	if r, _ := trainData.Size(); r == 0 {
		return errors.New("no training data. please increase ratio")
	}
	if r, _ := testData.Size(); r == 0 {
		return errors.New("no test data. please decrease ratio")
	}

	matrix, err := AssessWeatherClassification(trainData, testData)
	if err != nil {
		return fmt.Errorf("assess failed: %w", err)
	}

	_, r := testData.Size()
	w.Logger.Info("Accuracy", "rows", len(measurements), "testDats", r, "score", strconv.FormatFloat(evaluation.GetAccuracy(matrix), 'f', 3, 64))

	return nil
}

func removeUnknown(measurements []repository.Measurement) []repository.Measurement {
	var filtered []repository.Measurement
	for _, measurement := range measurements {
		if measurement.Weather != "UNKNOWN" {
			filtered = append(filtered, measurement)
		}
	}
	return filtered
}
