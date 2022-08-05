package server

import (
	"fmt"
	"net/http"
	"time"
)

func (server *Server) main(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`
<head>
  <meta http-equiv="Refresh" content="0; URL=/report">
</head>`))
}

func (server *Server) plot(w http.ResponseWriter, req *http.Request) {
	plotType, fold, start, stop, err := server.parseArgs(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = server.backend.PlotToWriter(plotType, fold, start, stop, w); err != nil {
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

	plotType = ScatterPlot
	if plotTypeString, found := req.URL.Query()["type"]; found {
		switch plotTypeString[0] {
		case "scatter":
			plotType = ScatterPlot
		case "contour":
			plotType = ContourPlot
		case "heatmap":
			plotType = HeatmapPlot
		default:
			err = fmt.Errorf("invalid plot type: %s", plotTypeString[0])
		}
	}
	return
}
