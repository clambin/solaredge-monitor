package plotters

import (
	"github.com/stretchr/testify/assert"
	"gonum.org/v1/plot/plotter"
	"testing"
)

func Test_makeGrid(t *testing.T) {
	const rows = 2
	const columns = 4
	var values = []float64{0, 1, 2, 3, 4, 5, 6, 7}

	var input plotter.XYZs
	for i := range values {
		// add 2 opposite values. the sum is always 7, so the average is always 3.5
		input = append(input, plotter.XYZ{X: float64(i % columns), Y: float64(i / columns), Z: values[i]})
		input = append(input, plotter.XYZ{X: float64(i % columns), Y: float64(i / columns), Z: values[len(values)-1-i]})
	}

	g := makeGrid(input, 2, 4)
	c, r := g.Dims()
	assert.Equal(t, columns, c)
	assert.Equal(t, rows, r)
	assert.Equal(t, []float64{0, 1, 2, 3}, g.xValues)
	assert.Equal(t, []float64{0, 1}, g.yValues)
	for row := range rows {
		for col := range columns {
			assert.Equal(t, 3.5, g.Z(col, row))
		}
	}
}
