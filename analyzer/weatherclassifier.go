package analyzer

import (
	"github.com/clambin/solaredge-monitor/store"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/knn"
	"strconv"
)

type WeatherClassifier struct {
	*knn.KNNClassifier
}

func NewWeatherClassifier() *WeatherClassifier {
	return &WeatherClassifier{
		KNNClassifier: knn.NewKnnClassifier("euclidean", "linear", 2),
	}
}

func AnalyzeMeasurements(measurements []store.Measurement) *base.DenseInstances {
	instances := base.NewDenseInstances()

	attribs := []base.Attribute{
		base.NewFloatAttribute("day"),
		base.NewFloatAttribute("timeOfDay"),
		base.NewFloatAttribute("intensity"),
		base.NewFloatAttribute("power"),
		base.NewFloatAttribute("weather"),
	}

	specs := make([]base.AttributeSpec, len(attribs))
	for i, a := range attribs {
		spec := instances.AddAttribute(a)
		specs[i] = spec
	}
	_ = instances.AddClassAttribute(attribs[len(attribs)-1])

	_ = instances.Extend(len(measurements))
	for i, measurement := range measurements {
		power, metrics := digitizeByPower(measurement)
		instances.Set(specs[0], i, specs[0].GetAttribute().GetSysValFromString(strconv.FormatFloat(metrics[0], 'f', 3, 64)))
		instances.Set(specs[1], i, specs[1].GetAttribute().GetSysValFromString(strconv.FormatFloat(metrics[1], 'f', 3, 64)))
		instances.Set(specs[2], i, specs[2].GetAttribute().GetSysValFromString(strconv.FormatFloat(metrics[2], 'f', 3, 64)))
		instances.Set(specs[3], i, specs[3].GetAttribute().GetSysValFromString(strconv.FormatFloat(power, 'f', 3, 64)))
		instances.Set(specs[4], i, specs[4].GetAttribute().GetSysValFromString(strconv.FormatFloat(metrics[3], 'f', 3, 64)))
	}
	return instances
}

func (w *WeatherClassifier) Classify(measurement store.Measurement) (string, error) {
	instances := AnalyzeMeasurements([]store.Measurement{measurement})
	result, err := w.Predict(instances)
	if err != nil {
		return "", err
	}
	attrib := result.AllAttributes()[4]
	spec, err := result.GetAttribute(attrib)
	if err != nil {
		return "", err
	}
	val := base.UnpackBytesToFloat(instances.Get(spec, 0))

	for name, id := range weatherIDs {
		if val == id {
			return name, nil
		}
	}
	return "???", nil
}
