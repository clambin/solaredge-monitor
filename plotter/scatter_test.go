package plotter_test

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/plotter"
	"github.com/clambin/solaredge-monitor/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/palette/moreland"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScatterPlotter_Plot(t *testing.T) {
	options := plotter.Options{
		Title: "foo",
		AxisX: plotter.Axis{
			Label:      "x",
			TimeFormat: "15:04:05",
		},
		AxisY: plotter.Axis{
			Label: "y",
		},
		Size: plotter.Size{
			Width:  800,
			Height: 600,
		},
		Legend: plotter.Legend{
			Increase: 500,
		},
		ColorMap: moreland.SmoothBlueRed(),
	}

	for i := 0; i < 2; i++ {
		var gpSuffix string
		fold := i == 0
		if fold {
			gpSuffix = "_folded"
		}

		p := plotter.ScatterPlotter{
			BasePlotter: plotter.BasePlotter{Options: options},
			Fold:        fold,
		}

		img, err := p.Plot(buildData(200))
		require.NoError(t, err)
		require.NotNil(t, img)

		buf := bytes.Buffer{}
		_, err = img.WriteTo(&buf)
		require.NoError(t, err)

		gp := filepath.Join("testdata", t.Name()+gpSuffix+"_golden.png")
		if *update {
			t.Logf("updating golden file for %s", t.Name())
			err = os.WriteFile(gp, buf.Bytes(), 0644)
			require.NoError(t, err, "failed to update golden file")
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.Equal(t, golden, buf.Bytes())
	}
}

func buildData(count int) (measurements []store.Measurement) {
	timestamp := time.Date(2022, time.July, 31, 0, 0, 0, 0, time.UTC)
	for hour := 0; hour < count; hour++ {
		var intensity float64
		hourOfDay := hour % 24
		if hourOfDay > 5 && hourOfDay < 21 {
			intensity = 100 * math.Sin((float64(hourOfDay)-5)*math.Pi/16)
		}
		measurements = append(measurements, store.Measurement{
			Timestamp: timestamp,
			Intensity: intensity,
			Power:     intensity * 40,
		})
		timestamp = timestamp.Add(time.Hour)
	}
	return measurements
}
