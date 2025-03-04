package plotters

import (
	"gonum.org/v1/plot/plotter"
	"math"
)

// we may want to split this in a base grid, which implements all of GridXYZ interface, except for Z().
// a MedianGrid could then get the sampler for Z, a MaxGrid could get the maximum, etc ...
// sampler would then become a sampler, which implements multiple function (median, average, min, max, etc.)

var _ plotter.GridXYZ = Grid{}

// A Grid groups the data of a plotter.XYZer in a plotter.GridXYZ, so it can be displayed as a Heat Map.
// For each cell, Z returns the *average* value.
type Grid struct {
	xValues []float64
	yValues []float64
	zValues []Sampler
	rows    int
	cols    int
}

func makeGrid(data plotter.XYZer, rows int, cols int) Grid {
	var g = Grid{rows: rows, cols: cols}

	// determine x & y ranges
	xMin, xMax, yMin, yMax := getXYRanges(data)
	xDelta := (xMax + 1 - xMin) / float64(cols)
	yDelta := (yMax + 1 - yMin) / float64(rows)

	g.xValues = makeRange(cols, xMin, xDelta)
	g.yValues = makeRange(rows, yMin, yDelta)

	// create z matrix
	g.zValues = make([]Sampler, rows*cols)
	for i := range data.Len() {
		x, y, z := data.XYZ(i)
		col := int((x - xMin) / xDelta)
		row := int((y - yMin) / yDelta)
		index := makeIndex(row, col, cols)
		g.zValues[index].Add(z)
	}

	return g
}

func makeIndex(row, col, cols int) int {
	return row*cols + col
}

func makeRange(count int, start, delta float64) []float64 {
	values := make([]float64, count)
	for i := range count {
		values[i] = start
		start += delta
	}
	return values
}

func getXYRanges(data plotter.XYZer) (float64, float64, float64, float64) {
	xMin, xMax := math.Inf(+1), math.Inf(-1)
	yMin, yMax := math.Inf(+1), math.Inf(-1)
	for i := range data.Len() {
		x, y := data.XY(i)
		if x < xMin {
			xMin = x
		} else if x > xMax {
			xMax = x
		}
		if y < yMin {
			yMin = y
		} else if y > yMax {
			yMax = y
		}
	}
	return xMin, xMax, yMin, yMax
}

func (g Grid) Dims() (c, r int) {
	return g.cols, g.rows
}

func (g Grid) X(c int) float64 {
	return g.xValues[c]
}

func (g Grid) Y(r int) float64 {
	return g.yValues[r]
}

func (g Grid) Z(c, r int) float64 {
	return g.zValues[makeIndex(r, c, g.cols)].Average()
}
