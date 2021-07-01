package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const ReportResponseTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>Report</title>
  </head>
  <body>
    <h1>Overview</h1>
    <h2>Time series</h2>
    <a href="/images/{{.TimeSeriesImage}}">
      <img src="/images/{{.TimeSeriesImage}}" alt="time series" style="width:400px;height:400px;">
    </a>
    <h2>Summary</h2>
    <a href="/images/{{.SummaryImage}}">
      <img src="/images/{{.SummaryImage}}" alt="summary" style="width:400px;height:400px;">
    </a>
  </body>
</html>`

func (server *Server) report(w http.ResponseWriter, req *http.Request) {
	start, stop, err := server.parseGenerateRequest(req)

	if err != nil {
		log.WithError(err).Error("failed to get determine start/stop parameters")
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer func() {
		if err != nil {
			log.WithError(err).Error("failed to create response")
			http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
		}
	}()

	var summary, timeSeries []byte
	if summary, err = server.backend.Summary(start, stop); err == nil {
		timeSeries, err = server.backend.TimeSeries(start, stop)
	}

	if err != nil {
		err = fmt.Errorf("failed to create report: %s", err.Error())
		return
	}

	var summaryFilename, timeSeriesFilename string
	if summaryFilename, err = server.cache.Store("summary.png", summary); err == nil {
		timeSeriesFilename, err = server.cache.Store("timeseries.png", timeSeries)
	}

	if err != nil {
		err = fmt.Errorf("failed to store image: %s", err.Error())
		return
	}

	data := struct {
		TimeSeriesImage string
		SummaryImage    string
	}{
		TimeSeriesImage: timeSeriesFilename,
		SummaryImage:    summaryFilename,
	}

	err = writePageFromTemplate(w, ReportResponseTemplate, data)

	if err != nil {
		err = fmt.Errorf("failed to create page from template: %s", err.Error())
	}
}
