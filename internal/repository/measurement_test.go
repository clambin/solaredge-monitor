package repository

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMeasurement_CompareTimestamp(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name  string
		other Measurement
		want  int
	}{
		{
			name:  "earlier",
			other: Measurement{Timestamp: now.Add(time.Hour)},
			want:  -1,
		},
		{
			name:  "same",
			other: Measurement{Timestamp: now},
			want:  0,
		},
		{
			name:  "later",
			other: Measurement{Timestamp: now.Add(-time.Hour)},
			want:  1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Measurement{Timestamp: now}
			assert.Equal(t, tt.want, m.CompareTimestamp(tt.other))
		})
	}
}
