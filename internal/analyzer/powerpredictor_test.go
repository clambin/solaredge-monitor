package analyzer

import (
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_digitizeByPower(t *testing.T) {
	tests := []struct {
		name  string
		args  repository.Measurement
		want  float64
		want1 []float64
	}{
		{
			name: "day 1",
			args: repository.Measurement{
				Timestamp: time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
				Power:     1000,
				Intensity: 55.0,
				Weather:   "SUN",
			},
			want:  1000,
			want1: []float64{1, 12, 55, 4},
		},
		{
			name: "mid year",
			args: repository.Measurement{
				Timestamp: time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC),
				Power:     1.0,
				Intensity: 5.0,
				Weather:   "SUN",
			},
			want:  1,
			want1: []float64{182, 0, 5, 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := digitizeByPower(tt.args)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, got1)
		})
	}
}
