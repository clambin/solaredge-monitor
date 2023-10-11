package analyzer

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/sjwhitworth/golearn/knn"
	"gonum.org/v1/gonum/mat"
	"slices"
)

func AssessPowerPrediction(trainingMeasurements, testMeasurements repository.Measurements) (repository.Measurements, float64, float64) {
	predicted := PredictPower(trainingMeasurements, testMeasurements)

	var variance, total float64
	for index, measurement := range testMeasurements {
		variance += measurement.Power - predicted[index].Power
		total += measurement.Power
	}

	pctVariance := variance / total
	variance /= float64(len(testMeasurements))
	return predicted, variance, pctVariance
}

func PredictPower(trainingMeasurements repository.Measurements, testMeasurements repository.Measurements) repository.Measurements {
	a := NewPowerPredictor(trainingMeasurements...)
	predicted := a.PredictSeries(testMeasurements)

	slices.SortFunc(predicted, func(a, b repository.Measurement) int {
		return a.CompareTimestamp(b)
	})

	return predicted
}

type PowerPredictor struct {
	regressor *knn.KNNRegressor
}

func NewPowerPredictor(measurements ...repository.Measurement) *PowerPredictor {
	p := PowerPredictor{
		regressor: knn.NewKnnRegressor("euclidean"),
	}

	if len(measurements) > 0 {
		p.Train(measurements)
	}

	return &p
}

const fieldCount = 4

func (p *PowerPredictor) Train(measurements []repository.Measurement) {
	output, input := createData(measurements)
	p.regressor.Fit(output, input, len(output), fieldCount)
}

func (p *PowerPredictor) Predict(measurement repository.Measurement) float64 {
	_, testData := createData([]repository.Measurement{measurement})

	testDataMatrix := mat.NewDense(1, fieldCount, testData)
	return p.regressor.Predict(testDataMatrix, 1)
}

func (p *PowerPredictor) PredictSeries(measurements []repository.Measurement) []repository.Measurement {
	predicted := make([]repository.Measurement, 0, len(measurements))
	const concurrentPredictions = 4
	input := make(chan repository.Measurement)
	output := make(chan repository.Measurement)

	go func() {
		for _, measurement := range measurements {
			input <- measurement
		}
		close(input)
	}()

	for i := 0; i < concurrentPredictions; i++ {
		go func() {
			for measurement := range input {
				measurement.Power = p.Predict(measurement)
				output <- measurement
			}
		}()
	}

	for i := 0; i < len(measurements); i++ {
		predicted = append(predicted, <-output)
	}

	slices.SortFunc(predicted, func(a, b repository.Measurement) int {
		return a.CompareTimestamp(b)
	})

	return predicted
}
