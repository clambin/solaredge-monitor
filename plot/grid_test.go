package plot_test

import (
	"github.com/clambin/solaredge-monitor/plot"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGridXYZ(t *testing.T) {
	xValues := []float64{1, 2, 3, 4}
	yValues := []float64{1, 2, 3, 4}
	zValues := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}

	g := plot.NewGrid(xValues, yValues, zValues)

	assert.Equal(t, 0.0, g.Min())
	assert.Equal(t, 1.0, g.Max())

	for index, x := range xValues {
		assert.Equal(t, x, g.X(index))
	}
	for index, y := range yValues {
		assert.Equal(t, y, g.Y(index))
	}
	for x := 0; x < len(xValues); x++ {
		for y := 0; y < len(yValues); y++ {
			if x == y {
				assert.Equal(t, 1.0, g.Z(x, y))
			} else {
				assert.Equal(t, 0.0, g.Z(x, y))
			}
		}
	}
}
