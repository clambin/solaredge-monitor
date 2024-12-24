package plotters

import (
	"bytes"
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestPlotXYZScatter(t *testing.T) {
	config := XYZConfig{
		Title:    "Report",
		X:        "time",
		XTicker:  "2006-01-02\n15:04:05",
		Y:        "solar intensity (%)",
		Width:    800,
		Height:   600,
		Ranges:   []float64{0, 1000, 2000, 3000, 4000},
		ColorMap: moreland.SmoothBlueRed(),
	}

	var output bytes.Buffer
	_, err := XYZScatter(&output, buildData(1000), config)
	require.NoError(t, err)

	gp := filepath.Join("testdata", strings.ToLower(t.Name()+"_golden.png"))
	if *update {
		require.NoError(t, os.WriteFile(gp, output.Bytes(), 0644))
	}
	golden, err := os.ReadFile(gp)
	require.NoError(t, err)
	assert.Equal(t, golden, output.Bytes())
}

func TestPlotXYZHeatmap(t *testing.T) {
	config := XYZConfig{
		Title:    "Report",
		X:        "time",
		XTicker:  "2006-01-02\n15:04:05",
		Y:        "solar intensity (%)",
		Width:    800,
		Height:   600,
		Ranges:   []float64{0, 1000, 2000, 3000, 4000},
		ColorMap: moreland.SmoothBlueRed(),
	}

	var output bytes.Buffer
	_, err := XYZHeatmap(&output, buildData(1000), config, 20, 24)
	require.NoError(t, err)

	gp := filepath.Join("testdata", strings.ToLower(t.Name())+"_golden.png")
	if *update {
		require.NoError(t, os.WriteFile(gp, output.Bytes(), 0644))
	}
	golden, err := os.ReadFile(gp)
	require.NoError(t, err)
	assert.Equal(t, golden, output.Bytes())
}

func buildData(count int) plotter.XYZs {
	data := make(plotter.XYZs, count)
	timestamp := time.Date(2022, time.July, 31, 0, 0, 0, 0, time.UTC)
	for hour := range count {
		var intensity float64
		hourOfDay := hour % 24
		if hourOfDay > 5 && hourOfDay < 21 {
			intensity = 100 * math.Sin((float64(hourOfDay)-5)*math.Pi/16)
		}
		data[hour] = plotter.XYZ{
			X: float64(timestamp.Unix()),
			Y: intensity,
			Z: intensity * 40,
		}
		timestamp = timestamp.Add(time.Hour)
	}
	return data
}
