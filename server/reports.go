package server

import (
	"bytes"
	"fmt"
	"github.com/clambin/solaredge-monitor/plotter"
	"github.com/clambin/solaredge-monitor/store"
	"gonum.org/v1/plot/palette/moreland"
	"gonum.org/v1/plot/vg/vgimg"
	"io"
	"time"
)

type Reporter struct {
	store.DB
}

type PlotType int

const (
	ScatterPlot = iota
	ContourPlot
	HeatmapPlot
)

func (r *Reporter) Plot(plotType PlotType, fold bool, start, stop time.Time) (image []byte, err error) {
	buf := new(bytes.Buffer)
	err = r.PlotToWriter(plotType, fold, start, stop, buf)
	return buf.Bytes(), err
}

func (r *Reporter) PlotToWriter(plotType PlotType, fold bool, start, stop time.Time, w io.Writer) (err error) {
	var measurements []store.Measurement
	if measurements, err = r.DB.Get(start, stop); err != nil {
		return
	}

	p := makePlotter(plotType, fold)
	var img *vgimg.PngCanvas
	if img, err = p.Plot(measurements); err != nil {
		return
	}

	_, err = img.WriteTo(w)
	return
}

const (
	width       = 800
	height      = 600
	xResolution = 48
	yResolution = 50
)

func makePlotter(plotType PlotType, fold bool) plotter.Plotter {
	timeFormat := "2006-01-02\n15:04:05"
	title := "Power output"
	if fold {
		timeFormat = "15:04:05"
		title = "Daily power output"
	}

	basePlotter := plotter.BasePlotter{
		Title:    title,
		AxisX:    plotter.Axis{Label: "time", TimeFormat: timeFormat},
		AxisY:    plotter.Axis{Label: "solar intensity (%)"},
		Size:     plotter.Size{Width: width, Height: height},
		ColorMap: moreland.SmoothBlueRed(),
		Fold:     fold,
	}

	var p plotter.Plotter
	switch plotType {
	case ScatterPlot:
		p = plotter.ScatterPlotter{
			BasePlotter: basePlotter,
			Legend:      plotter.Legend{Increase: 100},
		}
	case ContourPlot:
		p = plotter.ContourPlotter{GriddedPlotter: plotter.GriddedPlotter{
			BasePlotter: basePlotter,
			XResolution: xResolution,
			YResolution: yResolution,
			YRange:      &plotter.Range{Min: 0, Max: 110},
			Ranges:      []float64{1000, 2000, 3000, 3500, 3800, 4000},
		}}
	case HeatmapPlot:
		p = plotter.HeatmapPlotter{GriddedPlotter: plotter.GriddedPlotter{
			BasePlotter: basePlotter,
			XResolution: xResolution,
			YResolution: yResolution,
			YRange:      &plotter.Range{Min: 0, Max: 110},
			Ranges:      []float64{1000, 2000, 3000, 3500, 3800, 4000},
		}}
	}
	return p
}
func (r *Reporter) GetFirst() (first time.Time, err error) {
	var measurements []store.Measurement

	measurements, err = r.DB.GetAll()

	if err == nil && len(measurements) == 0 {
		err = fmt.Errorf("no entries found")
	}

	if err != nil {
		return
	}

	return measurements[0].Timestamp, nil
}

func (r *Reporter) GetLast() (first time.Time, err error) {
	var measurements []store.Measurement

	measurements, err = r.DB.GetAll()

	if err == nil && len(measurements) == 0 {
		err = fmt.Errorf("no entries found")
	}

	if err != nil {
		return
	}

	return measurements[len(measurements)-1].Timestamp, nil
}
