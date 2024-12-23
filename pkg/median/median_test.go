package median

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_median(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"odd number of values", []float64{1, 2, 3, 4, 0}, 2},
		{"even number of values", []float64{1, 2, 3, 4, 5, 0}, 2.5},
		{"empty slice", nil, 0.0},
		{"handle duplicates", []float64{1, 1, 1, 2}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Median
			m.Add(tt.values...)
			assert.Equal(t, tt.want, m.Median())

			m.Reset()
			assert.Zero(t, m.Median())
		})
	}
}
