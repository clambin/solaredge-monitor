package server

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func (server *Server) main(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte(`
<head>
  <meta http-equiv="Refresh" content="0; URL=/report">
</head>`))
}

func (server *Server) summary(w http.ResponseWriter, req *http.Request) {
	server.handleDetailRequest(w, req, "Summary", server.backend.Summary)
}

func (server *Server) timeSeries(w http.ResponseWriter, req *http.Request) {
	server.handleDetailRequest(w, req, "Time Series", server.backend.TimeSeries)
}

func (server *Server) classify(w http.ResponseWriter, req *http.Request) {
	server.handleDetailRequest(w, req, "Classification", server.backend.Classify)
}

const detailResponseTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
  </head>
  <body>
    <h1>{{.Title}}</h1>
    <a href="/images/{{.Filename}}">
      <img src="/images/{{.Filename}}" alt="summary" style="width:400px;height:400px;">
    </a>
  </body>
</html>`

func (server *Server) handleDetailRequest(w http.ResponseWriter, req *http.Request, title string, backendFunction func(time.Time, time.Time) ([]byte, error)) {
	start, stop, err := server.parseRequest(req)

	if err != nil {
		log.WithError(err).Error("failed to get determine start/stop parameters")
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	defer func() {
		if err != nil {
			log.WithError(err).Error("failed to create page")
			http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
		}
	}()

	var filename string
	filename, err = server.runReport(backendFunction, start, stop, "image.png")

	if err != nil {
		err = fmt.Errorf("failed to store image: %s", err.Error())
		return
	}

	data := struct {
		Title    string
		Filename string
	}{
		Title:    title,
		Filename: filename,
	}
	err = writePageFromTemplate(w, detailResponseTemplate, data)
}
