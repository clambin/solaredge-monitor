package analyzer_test

import (
	"github.com/clambin/solaredge-monitor/internal/analyzer"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/sjwhitworth/golearn/base"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAnalyzeMeasurements(t *testing.T) {
	measurements := []repository.Measurement{
		{
			Timestamp: time.Date(2023, time.January, 1, 12, 0, 0, 0, time.UTC),
			Power:     1000,
			Intensity: 55.0,
			Weather:   "SUN",
		},
	}

	data := analyzer.AnalyzeMeasurements(measurements)

	cols, rows := data.Size()
	require.Equal(t, 1, rows)
	require.Equal(t, 5, cols)

	attribs := data.AllClassAttributes()
	require.Len(t, attribs, 1)
	assert.Equal(t, "weather", attribs[0].GetName())

	tests := []struct {
		name string
		val  float64
	}{
		{name: "day", val: 1},
		{name: "timeOfDay", val: 12},
		{name: "intensity", val: 55},
		{name: "power", val: 1000},
		{name: "weather", val: 4},
	}

	attribs = data.AllAttributes()
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.name, attribs[i].GetName())
			spec, err := data.GetAttribute(attribs[i])
			require.NoError(t, err)
			assert.Equal(t, base.Float64Type, spec.GetAttribute().GetType())
			assert.Equal(t, tt.val, base.UnpackBytesToFloat(data.Get(spec, 0)))
		})
	}
}
