package plotters

import (
	"fmt"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"io"
)

type XYZConfig struct {
	Title    string
	X        string
	XTicker  string
	Y        string
	Width    float64
	Height   float64
	Ranges   []float64
	ColorMap palette.ColorMap
}

func XYZScatter(w io.Writer, data plotter.XYZer, config XYZConfig) (int64, error) {
	p := basePlot(config)
	sc, err := scatterPlot(data, config)
	if err != nil {
		return 0, err
	}
	p.Add(sc)
	addLegend(p, config)
	return writePng(w, p, config.Width, config.Height)
}

func XYZHeatmap(w io.Writer, data plotter.XYZer, config XYZConfig, rows, cols int) (int64, error) {
	p := basePlot(config)
	p.Add(heatmapPlot(data, config, rows, cols))
	// TODO: this legend doesn't really work with a heatmap: it maps the colors by the value range, but a heatmap decides its own colors.
	addLegend(p, config)
	return writePng(w, p, config.Width, config.Height)
}

func basePlot(config XYZConfig) *plot.Plot {
	p := plot.New()
	p.Title.Text = config.Title
	p.Title.Padding = vg.Centimeter
	p.X.Label.Text = config.X
	p.X.Tick.Marker = plot.TimeTicks{Format: config.XTicker}
	p.Y.Label.Text = config.Y
	p.Add(plotter.NewGrid())
	return p
}

func scatterPlot(data plotter.XYZer, config XYZConfig) (*plotter.Scatter, error) {
	sc, err := plotter.NewScatter(data)
	if err != nil {
		return nil, err
	}

	colors := config.ColorMap
	colors.SetMin(config.Ranges[0])
	colors.SetMax(config.Ranges[len(config.Ranges)-1])

	sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		_, _, z := data.XYZ(i)
		color, err := colors.At(z)
		if err != nil {
			if z < config.Ranges[0] {
				// At returns error if value is below min ...
				color, _ = colors.At(colors.Min())
			} else {
				// ... or if higher than maximum
				color, _ = colors.At(colors.Max())
			}
		}
		return draw.GlyphStyle{Color: color, Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
	}

	return sc, nil
}

func heatmapPlot(data plotter.XYZer, config XYZConfig, rows, cols int) *plotter.HeatMap {
	g := makeGrid(data, rows, cols)
	p := config.ColorMap.Palette(10 /*len(config.Ranges)*/)
	return plotter.NewHeatMap(g, p)
}

func addLegend(p *plot.Plot, config XYZConfig) {
	colorPalette := config.ColorMap.Palette(len(config.Ranges))
	thumbs := plotter.PaletteThumbnailers(colorPalette)
	for i := len(thumbs) - 1; i >= 0; i-- {
		t := thumbs[i]
		val := int(config.Ranges[i])
		p.Legend.Add(fmt.Sprintf("%d", val), t)
	}
	p.Legend.XOffs = legendWidth
}

const legendWidth = vg.Centimeter

func writePng(w io.Writer, p *plot.Plot, width float64, height float64) (int64, error) {
	rawImg := vgimg.New(vg.Points(width), vg.Points(height))
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return vgimg.PngCanvas{Canvas: rawImg}.WriteTo(w)
}
