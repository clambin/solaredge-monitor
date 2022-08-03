package server

import (
	log "github.com/sirupsen/logrus"
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
    <a href="/plot?type=scatter&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot?type=scatter&start={{.Start}}&stop={{.Stop}}" alt="scatter" style="width:400px;height:400px;">
    </a>
    <a href="/plot?type=scatter&fold=true&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot?type=scatter&fold=true&start={{.Start}}&stop={{.Stop}}" alt="scatter" style="width:400px;height:400px;">
    </a>
    <p>
    <a href="/plot?type=heatmap&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot?type=heatmap&start={{.Start}}&stop={{.Stop}}" alt="heatmap" style="width:400px;height:400px;">
    </a>
    <a href="/plot?type=heatmap&fold=true&start={{.Start}}&stop={{.Stop}}">
      <img src="/plot?type=heatmap&fold=true&start={{.Start}}&stop={{.Stop}}" alt="heatmap" style="width:400px;height:400px;">
    </a>
  </body>
</html>`

func (server *Server) report(w http.ResponseWriter, req *http.Request) {
	start, stop, err := server.parseRequest(req)

	if err != nil {
		log.WithError(err).Error("failed to get determine start/stop parameters")
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

	err = writePageFromTemplate(w, reportResponseTemplate, data)
	if err != nil {
		log.WithError(err).Error("failed to create response")
		http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
	}
}
