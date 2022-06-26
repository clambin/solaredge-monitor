package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func (server *Server) parseGenerateRequest(req *http.Request) (start, stop time.Time, err error) {

	start, err = parseTimestamp(req, "start", server.backend.GetFirst)

	if err == nil {
		stop, err = parseTimestamp(req, "stop", server.backend.GetLast)
	}

	if err == nil && stop.Before(start) {
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
		err = fmt.Errorf("invalid format for '%s' argument: %s", field, err.Error())
	}
	return
}

func writePageFromTemplate(w http.ResponseWriter, pageTemplate string, data interface{}) {
	var err error
	t := template.New("body")
	if t, err = t.Parse(pageTemplate); err == nil {
		err = t.Execute(w, data)
	}

	if err != nil {
		http.Error(w, "failed to create page: "+err.Error(), http.StatusInternalServerError)
	}
}
