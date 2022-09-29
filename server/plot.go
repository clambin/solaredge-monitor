package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func (s *Server) plot(w http.ResponseWriter, req *http.Request) {
	plotType, fold, start, stop, err := s.parseArgs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.backend.Plot(w, plotType, fold, start, stop)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var plotTypes = map[string]PlotType{
	"scatter": ScatterPlot,
	"contour": ContourPlot,
	"heatmap": HeatmapPlot,
}

func (s *Server) parseArgs(req *http.Request) (plotType PlotType, fold bool, start, stop time.Time, err error) {
	plotTypeString := mux.Vars(req)["type"]
	var found bool
	if plotType, found = plotTypes[plotTypeString]; !found {
		err = fmt.Errorf("invalid plot type: %s", plotTypeString)
		return
	}

	if start, err = parseTimestamp(req, "start", s.backend.GetFirst); err != nil {
		return
	}
	if stop, err = parseTimestamp(req, "stop", s.backend.GetLast); err != nil {
		return
	}

	if stop.Before(start) {
		err = fmt.Errorf("start time is later than stop time")
		return
	}

	var foldString []string
	if foldString, found = req.URL.Query()["fold"]; found {
		fold = foldString[0] == "true"
	}

	return
}
