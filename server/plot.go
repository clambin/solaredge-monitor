package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

func (s *Server) plot(w http.ResponseWriter, req *http.Request) {
	plotType, fold, start, stop, err := s.parsePlotRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = s.backend.Plot(w, plotType, fold, start, stop); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var plotTypes = map[string]PlotType{
	"scatter": ScatterPlot,
	"contour": ContourPlot,
	"heatmap": HeatmapPlot,
}

func (s *Server) parsePlotRequest(req *http.Request) (plotType PlotType, fold bool, start, stop time.Time, err error) {
	plotTypeString := chi.URLParam(req, "type")
	var found bool
	if plotType, found = plotTypes[plotTypeString]; !found {
		err = fmt.Errorf("invalid plot type: %s", plotTypeString)
		return
	}

	if start, stop, err = s.parseTimestamps(req); err != nil {
		err = fmt.Errorf("timestamps: %w", err)
		return
	}

	if foldString, ok := req.URL.Query()["fold"]; ok {
		fold = foldString[0] == "true"
	}

	return
}
