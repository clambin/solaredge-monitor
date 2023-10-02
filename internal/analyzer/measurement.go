package analyzer

import "github.com/clambin/solaredge-monitor/internal/repository"

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
