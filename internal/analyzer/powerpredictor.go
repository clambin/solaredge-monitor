package analyzer

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/sjwhitworth/golearn/knn"
	"gonum.org/v1/gonum/mat"
	"sort"
)

type PowerPredictor struct {
	regressor *knn.KNNRegressor
}

func NewPowerPredictor() *PowerPredictor {
	return &PowerPredictor{
		regressor: knn.NewKnnRegressor("euclidean"),
	}
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
	output := make(chan repository.Measurement, concurrentPredictions)

	for i := 0; i < concurrentPredictions; i++ {
		go func() {
			for measurement := range input {
				measurement.Power = p.Predict(measurement)
				output <- measurement
			}
		}()
	}

	go func() {
		for _, measurement := range measurements {
			input <- measurement
		}
		close(input)
	}()

	for i := 0; i < len(measurements); i++ {
		predicted = append(predicted, <-output)
	}
	sort.Slice(predicted, func(i, j int) bool { return predicted[i].Timestamp.Before(predicted[j].Timestamp) })
	return predicted
}

func createData(measurements []repository.Measurement) ([]float64, []float64) {
	power := make([]float64, 0, len(measurements))
	input := make([]float64, 0, fieldCount*len(measurements))

	for _, measurement := range measurements {
		p, d := digitizeByPower(measurement)
		power = append(power, p)
		input = append(input, d...)
	}

	return power, input
}

func digitizeByPower(measurement repository.Measurement) (float64, []float64) {
	return measurement.Power, []float64{
		float64(measurement.Timestamp.YearDay()),
		float64(measurement.Timestamp.Hour()) + float64(measurement.Timestamp.Minute())/60 + float64(measurement.Timestamp.Second()/3600),
		measurement.Intensity,
		weatherID(measurement.Weather),
	}
}

var weatherIDs = map[string]float64{
	"NIGHT_CLOUDY":   2,
	"NIGHT_CLEAR":    3,
	"SUN":            4,
	"CLOUDY_MOSTLY":  5,
	"CLOUDY":         6,
	"CLOUDY_PARTLY":  7,
	"SCATTERED_RAIN": 8,
	"UNKNOWN":        1,
	"DRIZZLE":        9,
	"RAIN":           10,
	"SCATTER_SNOW":   11,
	"FOGGY":          12,
	"SNOW":           13,
	"THUNDERSTORM":   14,
}

func weatherID(weather string) float64 {
	return weatherIDs[weather]
}
