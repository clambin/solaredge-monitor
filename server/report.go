package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
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
	start, stop, err := s.parseRequest(req)

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

func (s *Server) parseRequest(req *http.Request) (start, stop time.Time, err error) {
	if start, err = parseTimestamp(req, "start", s.backend.GetFirst); err != nil {
		return
	}

	if stop, err = parseTimestamp(req, "stop", s.backend.GetLast); err != nil {
		return
	}

	if stop.Before(start) {
		err = fmt.Errorf("start time is later than stop time")
	}

	return
}

func parseTimestamp(req *http.Request, field string, dbfunc func() (time.Time, error)) (timestamp time.Time, err error) {
	arg, ok := req.URL.Query()[field]

	if !ok {
		return dbfunc()
	}

	timestamp, err = time.Parse(time.RFC3339, arg[0])

	if err != nil {
		err = fmt.Errorf("invalid format for '%s': %w", field, err)
	}
	return
}

func writePageFromTemplate(w http.ResponseWriter, pageTemplate string, data interface{}) (err error) {
	t := template.New("body")
	if t, err = t.Parse(pageTemplate); err == nil {
		err = t.Execute(w, data)
	}
	return
}
