package plot

import (
	"fmt"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/vgimg"
	"math"
)

type Options struct {
	Title   string
	AxisX   Axis
	AxisY   Axis
	Legend  Legend
	Size    Size
	Contour Contour
}

type Axis struct {
	Label      string
	TimeFormat string
}

type Legend struct {
	Increase int
}

type Size struct {
	Width  int
	Height int
}

type Contour struct {
	Ranges []float64
}

func ScatterPlot(data plotter.XYZs, options Options) (img *vgimg.PngCanvas, err error) {
	minZ, maxZ := getMinMax(data)

	p := makeBasePlot(options)
	c := allocateColors(minZ, maxZ)

	if err = addScatter(p, c, data, minZ, maxZ); err == nil {
		if minZ != math.Inf(+1) && maxZ != math.Inf(-1) {
			increase := int(math.Max(1, float64(options.Legend.Increase)))
			addLegend(p, c, minZ, maxZ, increase)
		}
		img = saveImage(p, options)
	}
	return
}

func ContourPlot(data *GridXYZ, options Options) (img *vgimg.PngCanvas, err error) {
	rows, cols := data.Dims()
	if rows == 0 || cols == 0 {
		return nil, fmt.Errorf("no data to plot")
	}
	p := makeBasePlot(options)

	if err = addContour(p, data, options.Contour); err == nil {
		if len(options.Contour.Ranges) > 0 {
			_ = addContourLegend(p, options.Contour.Ranges)
		}
		img = saveImage(p, options)
	}
	return
}
