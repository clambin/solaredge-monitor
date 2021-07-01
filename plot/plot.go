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
	if data.Len() == 0 {
		return nil, fmt.Errorf("no data to plot")
	}
	minZ, maxZ := getMinMax(data)
	p := makeBasePlot(options)
	c := allocateColors(minZ, maxZ)

	if err = addScatter(p, c, data, minZ, maxZ); err == nil {
		increase := int(math.Max(1, float64(options.Legend.Increase)))

		makeLegend(p, c, minZ, maxZ, increase)
		img = saveImage(p, options)
	}
	return
}

func ContourPlot(data *GridXYZ, options Options) (img *vgimg.PngCanvas, err error) {
	rows, cols := data.Dims()
	if rows == 0 || cols == 0 {
		return nil, fmt.Errorf("no data to plot")
	}
	minZ, maxZ := data.Min(), data.Max()
	p := makeBasePlot(options)
	c := allocateColors(minZ, maxZ)

	if err = addContour(p, c, data, options.Contour); err == nil {
		// increase := int(math.Max(10, float64(options.Legend.Increase)))
		// makeLegend(p, c, minZ, maxZ, increase)
		img = saveImage(p, options)
	}
	return
}
