package plotter_test

import (
	"bytes"
	"github.com/clambin/solaredge-monitor/internal/web/plotter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestContourPlotter_Plot(t *testing.T) {
	for _, fold := range []bool{true, false} {
		var gpSuffix string
		if fold {
			gpSuffix = "_folded"
		}

		p := plotter.ContourPlotter{
			GriddedPlotter: plotter.NewGriddedPlotter("foo"),
		}

		var buf bytes.Buffer
		img, err := p.Plot(buildData(200), fold)
		require.NoError(t, err)
		_, err = img.WriteTo(&buf)
		require.NoError(t, err)
		require.NotZero(t, buf.Len())

		gp := filepath.Join("testdata", t.Name()+gpSuffix+"_golden.png")
		if *update {
			err = os.WriteFile(gp, buf.Bytes(), 0644)
			require.NoError(t, err)
		}

		var golden []byte
		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(golden, buf.Bytes()))
	}
}
