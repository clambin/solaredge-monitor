package plot

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"math"
)

func getMinMax(data plotter.XYZs) (min, max float64) {
	min = math.Inf(+1)
	max = math.Inf(-1)
	for i := 0; i < data.Len(); i++ {
		_, _, z := data.XYZ(i)
		min = math.Min(min, z)
		max = math.Max(max, z)
	}
	return
}

func makeBasePlot(options Options) (p *plot.Plot) {
	p = plot.New()
	p.Title.Text = options.Title
	p.X.Label.Text = options.AxisX.Label
	if options.AxisX.TimeFormat != "" {
		p.X.Tick.Marker = plot.TimeTicks{Format: options.AxisX.TimeFormat}
	}
	p.Y.Label.Text = options.AxisY.Label
	p.Add(plotter.NewGrid())

	return p
}

func allocateColors(minZ, maxZ float64) palette.ColorMap {
	colors := moreland.SmoothBlueRed() // Initialize a color map.
	colors.SetMin(minZ)
	colors.SetMax(maxZ)

	return colors
}

// This is the width of the legend, experimentally determined.
const legendWidth = vg.Centimeter

func addLegend(p *plot.Plot, c palette.ColorMap, minZ, maxZ float64, increase int) {
	step := int(maxZ-minZ) / increase
	thumbs := plotter.PaletteThumbnailers(c.Palette(step))
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		if i != 0 && i != len(thumbs)-1 {
			p.Legend.Add("", t)
			continue
		}
		var val int
		switch i {
		case 0:
			val = int(minZ)
		case len(thumbs) - 1:
			val = int(maxZ)
		}
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}

	p.Legend.XOffs = legendWidth
}

func saveImage(p *plot.Plot, options Options) *vgimg.PngCanvas {
	rawImg := vgimg.New(vg.Points(float64(options.Size.Width)), vg.Points(float64(options.Size.Height)))
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return &vgimg.PngCanvas{Canvas: rawImg}
}

func addScatter(p *plot.Plot, c palette.ColorMap, data plotter.XYZs, minZ, maxZ float64) (err error) {
	var sc *plotter.Scatter
	sc, err = plotter.NewScatter(data)

	if err == nil {
		sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
			_, _, z := data.XYZ(i)
			d := (z - minZ) / (maxZ - minZ)
			rng := maxZ - minZ
			k := d*rng + minZ
			color, _ := c.At(k)
			return draw.GlyphStyle{Color: color, Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
		}
		p.Add(sc)
	}
	return
}

func addContour(p *plot.Plot, data *GridXYZ, contour Contour) (err error) {
	colorCount := len(contour.Ranges)
	if colorCount == 0 {
		colorCount = 5
	}
	colors := palette.Rainbow(colorCount, palette.Blue, palette.Red, 1, 1, 1)

	ct := plotter.NewContour(data, contour.Ranges, colors)
	p.Add(ct)

	return
}

func addContourLegend(p *plot.Plot, ranges []float64) (err error) {
	colors := palette.Rainbow(len(ranges), palette.Blue, palette.Red, 1, 1, 1)
	thumbs := plotter.PaletteThumbnailers(colors)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		val := int(ranges[i])
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}

	p.Legend.XOffs = legendWidth

	return
}
