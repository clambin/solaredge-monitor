package server

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func (server *Server) plot(w http.ResponseWriter, req *http.Request) {
	plotType, fold, start, stop, err := server.parseArgs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = server.backend.PlotToWriter(plotType, fold, start, stop, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (server *Server) parseArgs(req *http.Request) (plotType PlotType, fold bool, start, stop time.Time, err error) {
	if start, err = parseTimestamp(req, "start", server.backend.GetFirst); err != nil {
		return
	}
	if stop, err = parseTimestamp(req, "stop", server.backend.GetLast); err != nil {
		return
	}

	if stop.Before(start) {
		err = fmt.Errorf("start time is later than stop time")
		return
	}

	if foldString, found := req.URL.Query()["fold"]; found {
		fold = foldString[0] == "true"
	}

	v := mux.Vars(req)
	switch v["type"] {
	case "scatter":
		plotType = ScatterPlot
	case "contour":
		plotType = ContourPlot
	case "heatmap":
		plotType = HeatmapPlot
	default:
		err = fmt.Errorf("invalid plot type: %s", v["type"])
	}

	return
}
