package plotter_test

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/plotter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/plot/palette/moreland"
	"os"
	"path/filepath"
	"testing"
)

func TestHeatmapPlotter_Plot(t *testing.T) {
	options := plotter.Options{
		Title: "foo",
		AxisX: plotter.Axis{
			Label:      "time",
			TimeFormat: "15:04:05",
		},
		AxisY: plotter.Axis{
			Label: "intensity (%)",
		},
		Size: plotter.Size{
			Width:  800,
			Height: 600,
		},
		ColorMap: moreland.SmoothBlueRed(),
		Contour: plotter.Contour{
			Ranges: []float64{1000, 2000, 3000, 4000},
		},
	}

	for i := 0; i < 2; i++ {
		var gpSuffix string
		fold := i == 0
		if fold {
			gpSuffix = "_folded"
		}

		p := plotter.HeatmapPlotter{
			BasePlotter: plotter.BasePlotter{Options: options},
			Fold:        fold,
			XSteps:      48,
			YSteps:      10,
			YRange:      plotter.NewRange(0, 100),
		}

		img, err := p.Plot(buildData(200))
		assert.NoError(t, err)
		assert.NotNil(t, img)

		buf := bytes.Buffer{}
		_, err = img.WriteTo(&buf)
		require.NoError(t, err)

		gp := filepath.Join("testdata", t.Name()+gpSuffix+"_golden.png")
		if *update {
			err = os.WriteFile(gp, buf.Bytes(), 0644)
			require.NoError(t, err)
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.Equal(t, golden, buf.Bytes())
	}
}
