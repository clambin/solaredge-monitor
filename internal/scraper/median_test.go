package scraper

import (
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func Test_median(t *testing.T) {
	tests := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"odd number of values", []float64{0, 1, 2, 3, 4}, 2},
		{"even number of values", []float64{0, 1, 2, 3, 4, 5}, 2.5},
		{"empty slice", nil, 0.0},
		{"handle duplicates", []float64{1, 1, 1, 2}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m median
			for _, value := range slices.Backward(tt.values) {
				m.add(value)
			}
			assert.Equal(t, tt.want, m.median())
			assert.Equal(t, 0.0, m.median())
		})
	}
}
