package plot_test

import (
	"github.com/clambin/solaredge-monitor/plot"
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/plot/plotter"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestGraph_ScatterPlot(t *testing.T) {
	options := plot.Options{
		Title: "foo",
		AxisX: plot.Axis{
			Label:      "x",
			TimeFormat: "15:04:05",
		},
		AxisY: plot.Axis{
			Label: "y",
		},
		Size: plot.Size{
			Width:  800,
			Height: 600,
		},
	}
	img, err := plot.ScatterPlot(buildData(200), options)
	assert.NoError(t, err)
	assert.NotNil(t, img)
}

func TestGraph_ContourPlot(t *testing.T) {
	options := plot.Options{
		Title: "foo",
		AxisX: plot.Axis{
			Label: "x",
		},
		AxisY: plot.Axis{
			Label: "y",
		},
		Size: plot.Size{
			Width:  800,
			Height: 600,
		},
		Contour: plot.Contour{
			Ranges: []float64{},
		},
	}
	img, err := plot.ContourPlot(buildGridData(20, 20), options)
	assert.NoError(t, err)
	assert.NotNil(t, img)

	var w *os.File
	w, err = os.Create("foo.png")
	assert.NoError(t, err)
	_, err = img.WriteTo(w)
	assert.NoError(t, err)
	_ = w.Close()

}

func buildData(size int) (data plotter.XYZs) {
	data = make(plotter.XYZs, size)
	for i := 0; i < size; i++ {
		data[i].X = rand.Float64() - 0.5
		data[i].Y = rand.Float64() - 0.5
		data[i].Z = rand.Float64() - 0.5
	}
	return
}

func buildGridData(c, r int) (data *plot.GridXYZ) {
	rand.Seed(time.Now().Unix())
	X := make([]float64, c)
	for i := 0; i < c; i++ {
		X[i] = float64(i)
	}

	Y := make([]float64, r)
	for i := 0; i < r; i++ {
		Y[i] = float64(i)
	}

	Z := make([]float64, c*r)
	for i := 0; i < c*r; i++ {
		Z[i] = float64(i)*rand.Float64() + float64(i%c) - 0.5
	}
	return plot.NewGrid(X, Y, Z)
}
