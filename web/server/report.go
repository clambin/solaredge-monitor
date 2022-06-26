package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

const reportResponseTemplate = `
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
    <h2>Classification</h2>
    <a href="/images/{{.ClassifyImage}}">
      <img src="/images/{{.ClassifyImage}}" alt="classification" style="width:400px;height:400px;">
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

	var wg sync.WaitGroup
	wg.Add(3)

	var summaryFilename, timeSeriesFilename, classificationFilename string
	errs := make([]error, 3)

	go func() {
		summaryFilename, errs[0] = server.runReport(server.backend.Summary, start, stop, "summary.png")
		wg.Done()
	}()

	go func() {
		timeSeriesFilename, errs[1] = server.runReport(server.backend.TimeSeries, start, stop, "timeseries.png")
		wg.Done()
	}()

	go func() {
		classificationFilename, errs[2] = server.runReport(server.backend.Classify, start, stop, "classify.png")
		wg.Done()
	}()

	wg.Wait()
	for _, e := range errs {
		if e != nil {
			err = e
			return
		}
	}

	data := struct {
		TimeSeriesImage string
		SummaryImage    string
		ClassifyImage   string
	}{
		TimeSeriesImage: timeSeriesFilename,
		SummaryImage:    summaryFilename,
		ClassifyImage:   classificationFilename,
	}

	err = writePageFromTemplate(w, reportResponseTemplate, data)
}

func (server *Server) runReport(f func(start time.Time, stop time.Time) (image []byte, err error), start, stop time.Time, name string) (filename string, err error) {
	var img []byte
	if img, err = f(start, stop); err == nil {
		filename, err = server.cache.Store(name, img)
	}

	if err != nil {
		err = fmt.Errorf("report %s failed: %w", name, err)
	}

	return
}
