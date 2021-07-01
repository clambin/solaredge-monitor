package plot

import (
	"gonum.org/v1/gonum/mat"
	"math"
)

type GridXYZ struct {
	x   []float64
	y   []float64
	z   mat.Matrix
	min float64
	max float64
}

func NewGrid(x, y []float64, z []float64) *GridXYZ {
	min := math.Inf(+1)
	max := math.Inf(-1)
	for _, zVal := range z {
		min = math.Min(min, zVal)
		max = math.Max(max, zVal)
	}

	return &GridXYZ{
		x:   x,
		y:   y,
		z:   mat.NewDense(len(x), len(y), z),
		min: min,
		max: max,
	}
}

func (g GridXYZ) Dims() (c, r int) {
	return len(g.x), len(g.y)
}

func (g GridXYZ) Z(c, r int) float64 {
	return g.z.At(c, r)
}

func (g GridXYZ) X(c int) float64 {
	return g.x[c]
}
func (g GridXYZ) Y(r int) float64 {
	return g.y[r]
}

func (g GridXYZ) Min() float64 {
	return g.min
}

func (g GridXYZ) Max() float64 {
	return g.max
}
