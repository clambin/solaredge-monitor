package server

import (
	"golang.org/x/exp/slog"
	"html/template"
	"net/http"
	"time"
)

const reportResponseTemplate = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Report</title>
  </head>
  <body>
    <a href="/plot/scatter?start={{.Start}}&stop={{.Stop}}">
      <img src="/plot/scatter?start={{.Start}}&stop={{.Stop}}" alt="scatter" style="width:400px;height:400px;">
    </a>
    <a href="/plot/scatter?fold=true&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot/scatter?fold=true&start={{.Start}}&stop={{.Stop}}" alt="scatter" style="width:400px;height:400px;">
    </a>
    <p>
    <a href="/plot/heatmap?start={{.Start}}&stop={{.Stop}}">
      <img src="/plot/heatmap?start={{.Start}}&stop={{.Stop}}" alt="heatmap" style="width:400px;height:400px;">
    </a>
    <a href="/plot/heatmap?fold=true&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot/heatmap?fold=true&start={{.Start}}&stop={{.Stop}}" alt="heatmap" style="width:400px;height:400px;">
    </a>
  </body>
</html>`

func (s *Server) report(w http.ResponseWriter, req *http.Request) {
	start, stop, err := s.parseTimestamps(req)
	if err != nil {
		slog.Error("failed to determine start/stop parameters", "err", err)
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	data := struct {
		Start string
		Stop  string
	}{
		Start: start.Format(time.RFC3339),
		Stop:  stop.Format(time.RFC3339),
	}

	t := template.New("body")
	if t, err = t.Parse(reportResponseTemplate); err == nil {
		err = t.Execute(w, data)
	}

	if err != nil {
		slog.Error("failed to create response", "err", err)
		http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
	}
}
