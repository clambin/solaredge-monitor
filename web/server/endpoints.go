package server

import (
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

const DetailResponseTemplate = `
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
	start, stop, err := server.parseGenerateRequest(req)

	if err != nil {
		log.WithError(err).Error("failed to get determine start/stop parameters")
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	var output []byte

	if output, err = backendFunction(start, stop); err != nil {
		log.WithError(err).Error("failed to create image")
		http.Error(w, "unable to create image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var filename string
	if filename, err = server.cache.Store("image.png", output); err != nil {
		log.WithError(err).Error("failed to store images")
		http.Error(w, "unable to store image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		Filename string
	}{
		Title:    title,
		Filename: filename,
	}

	err = writePageFromTemplate(w, DetailResponseTemplate, data)

	if err != nil {
		http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
