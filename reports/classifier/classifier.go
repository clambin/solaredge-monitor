package classifier

import (
	"github.com/clambin/solaredge-monitor/store"
	nn "github.com/pa-m/sklearn/neural_network"
	"github.com/pa-m/sklearn/preprocessing"
	"gonum.org/v1/gonum/mat"
)

type Classifier struct {
	classifier *nn.MLPClassifier
	scaler     *preprocessing.StandardScaler
}

func New(batchSize, maxIter int) *Classifier {
	var (
		hiddenLayers = []int{100}
		Alpha        float64
	)
	c := nn.NewMLPClassifier(hiddenLayers, "relu", "adam", Alpha)
	c.BatchSize = batchSize
	c.MaxIter = maxIter
	return &Classifier{
		classifier: c,
		scaler:     preprocessing.NewStandardScaler(),
	}
}

func (classifier *Classifier) Learn(measurements []store.Measurement) {
	trainX, trainY := getXY(measurements)
	classifier.scaler.Fit(trainX, trainY)
	scaledX, scaledY := classifier.scaler.Transform(trainX, trainY)
	classifier.classifier.Fit(scaledX, scaledY)
}

func (classifier *Classifier) Score(measurements []store.Measurement) float64 {
	testX, testY := getXY(measurements)
	classifier.scaler.Fit(testX, testY)
	scaledX, scaledY := classifier.scaler.Transform(testX, testY)

	return classifier.classifier.Score(scaledX, scaledY)
}

func (classifier *Classifier) Classify(measurements []store.Measurement) (classified []store.Measurement) {
	X, Y := getXY(measurements)
	classifier.scaler.Fit(X, Y)
	scaledX, _ := classifier.scaler.Transform(X, Y)

	var predictions mat.Dense
	fc := classifier.classifier.Predict(scaledX, &predictions)

	_, unscaledY := classifier.scaler.InverseTransform(scaledX, fc)

	for index, measurement := range measurements {
		classified = append(classified, store.Measurement{
			Timestamp: measurement.Timestamp,
			Intensity: measurement.Intensity,
			Power:     unscaledY.At(index, 0),
		})
	}

	return
}

func getXY(measurements []store.Measurement) (X, Y *mat.Dense) {
	dataX := make([]float64, 2*len(measurements))
	dataY := make([]float64, len(measurements))

	for index, measurement := range measurements {
		dataX[2*index] = float64(measurement.Timestamp.Unix() % (24 * 60 * 60))
		dataX[2*index+1] = measurement.Intensity
		dataY[index] = measurement.Power
	}

	X = mat.NewDense(len(measurements), 2, dataX)
	Y = mat.NewDense(len(measurements), 1, dataY)

	return
}
