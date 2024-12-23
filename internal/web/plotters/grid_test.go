package plotters

import (
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/plot/plotter"
	"testing"
)

func Test_makeGrid(t *testing.T) {
	const rows = 2
	const columns = 4
	var xyzs plotter.XYZs
	value := float64(1)
	for row := range rows {
		for col := range columns {
			xyzs = append(xyzs, plotter.XYZ{X: float64(col), Y: float64(row), Z: value})
			value++
		}
	}

	g := makeGrid(xyzs, 2, 4)
	c, r := g.Dims()
	assert.Equal(t, columns, c)
	assert.Equal(t, rows, r)
	assert.Equal(t, []float64{0, 1, 2, 3}, g.xValues)
	assert.Equal(t, []float64{0, 1}, g.yValues)
	want := float64(1)
	for row := range rows {
		for col := range columns {
			assert.Equal(t, want, g.Z(col, row))
			want++
		}
	}
}
