package plotter_test

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/internal/repository"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScatterPlotter_Plot(t *testing.T) {
	for _, fold := range []bool{true, false} {
		var gpSuffix string
		if fold {
			gpSuffix = "_folded"
		}

		p := plotter.ScatterPlotter{
			BasePlotter: plotter.NewBasePlotter("foo"),
			Legend:      plotter.Legend{Increase: 100},
		}

		var buf bytes.Buffer
		img, err := p.Plot(buildData(200), fold)
		assert.NoError(t, err)
		_, err = img.WriteTo(&buf)
		assert.NoError(t, err)
		assert.NotZero(t, buf.Len())

		gp := filepath.Join("testdata", t.Name()+gpSuffix+"_golden.png")
		if *update {
			t.Logf("updating golden file for %s", t.Name())
			err = os.WriteFile(gp, buf.Bytes(), 0644)
			require.NoError(t, err, "failed to update golden file")
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(golden, buf.Bytes()))
	}
}

func buildData(count int) (measurements repository.Measurements) {
	timestamp := time.Date(2022, time.July, 31, 0, 0, 0, 0, time.UTC)
	for hour := range count {
		var intensity float64
		hourOfDay := hour % 24
		if hourOfDay > 5 && hourOfDay < 21 {
			intensity = 100 * math.Sin((float64(hourOfDay)-5)*math.Pi/16)
		}
		measurements = append(measurements, repository.Measurement{
			Timestamp: timestamp,
			Intensity: intensity,
			Power:     intensity * 40,
		})
		timestamp = timestamp.Add(time.Hour)
	}
	return measurements
}
