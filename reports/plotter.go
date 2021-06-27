package reports

import (
	"fmt"
	"github.com/clambin/solaredge-monitor/store"
	log "github.com/sirupsen/logrus"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"
	"math"
)

// This is the width of the legend, experimentally determined.
const legendWidth = vg.Centimeter

func MakeGraph(measurements []store.Measurement, fold bool) (img *vgimg.PngCanvas, err error) {
	if len(measurements) == 0 {
		return nil, fmt.Errorf("no data for graph")
	}

	XYZs, minZ, maxZ := buildPlotData(measurements, fold)

	var p *plot.Plot
	if p, err = makePlot(XYZs, minZ, maxZ, fold); err != nil {
		return nil, err
	}

	return makeImage(p), nil
}

func buildPlotData(input []store.Measurement, fold bool) (result plotter.XYZs, minZ, maxZ float64) {
	result = make(plotter.XYZs, len(input))

	minZ, maxZ = math.Inf(1), math.Inf(-1)
	index := 0
	for _, value := range input {
		unixTime := float64(value.Timestamp.Unix())
		if fold {
			unixTime = float64(int64(unixTime) % (60 * 60 * 24))
		}
		result[index].X = unixTime
		result[index].Y = value.Intensity
		Z := value.Power
		result[index].Z = Z

		if Z < minZ {
			minZ = Z
		}
		if Z > maxZ {
			maxZ = Z
		}
		index++
	}
	return
}

func makePlot(XYZs plotter.XYZs, minZ, maxZ float64, fold bool) (p *plot.Plot, err error) {
	colors := moreland.SmoothBlueRed() // Initialize a color map.
	colors.SetMax(maxZ)
	colors.SetMin(minZ)

	p = plot.New()
	p.Title.Text = "Solar Panel output"
	p.X.Label.Text = "Time"
	if fold {
		p.X.Tick.Marker = plot.TimeTicks{Format: "15:04:05"}
	} else {
		p.X.Tick.Marker = plot.TimeTicks{Format: "2006-01-02"}
	}
	p.Y.Label.Text = "Solar Intensity (%)"
	p.Add(plotter.NewGrid())

	var sc *plotter.Scatter
	sc, err = plotter.NewScatter(XYZs)

	if err != nil {
		log.WithError(err).Error("failed to create plotter")
		return nil, err
	}

	// Specify style and color for individual points.
	sc.GlyphStyleFunc = func(i int) draw.GlyphStyle {
		_, _, z := XYZs.XYZ(i)
		d := (z - minZ) / (maxZ - minZ)
		rng := maxZ - minZ
		k := d*rng + minZ
		c, _ := colors.At(k)
		return draw.GlyphStyle{Color: c, Radius: vg.Points(3), Shape: draw.CircleGlyph{}}
	}
	p.Add(sc)

	//Create a legend
	step := int(maxZ-minZ) / 100
	thumbs := plotter.PaletteThumbnailers(colors.Palette(step))
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

	// Slide the legend over so it doesn't overlap the ScatterPlot.
	p.Legend.XOffs = legendWidth

	return
}

func makeImage(p *plot.Plot) (img *vgimg.PngCanvas) {
	rawImg := vgimg.New(600, 460)
	dc := draw.New(rawImg)
	dc = draw.Crop(dc, 0, -legendWidth, 0, 0) // Make space for the legend.
	p.Draw(dc)

	return &vgimg.PngCanvas{Canvas: rawImg}
}
