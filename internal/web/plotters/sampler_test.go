package plotters

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_median(t *testing.T) {
	tests := []struct {
		name    string
		values  []float64
		median  float64
		average float64
		min     float64
		max     float64
	}{
		{"odd number of values", []float64{1, 2, 3, 4, 0}, 2, 2, 0, 4},
		{"even number of values", []float64{1, 2, 3, 4, 5, 0}, 2.5, 2.5, 0, 5},
		{"empty slice", nil, 0.0, 0, 0, 0},
		{"handle duplicates", []float64{1, 1, 1, 2}, 1, 5.0 / 4, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Sampler
			m.Add(tt.values...)
			assert.Equal(t, tt.median, m.Median())
			assert.Equal(t, tt.average, m.Average())
			assert.Equal(t, tt.min, m.Min())
			assert.Equal(t, tt.max, m.Max())

			m.Reset()
			assert.Zero(t, m.Median())
		})
	}
}
