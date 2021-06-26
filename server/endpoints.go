package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"
)

func (server *Server) main(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

const OverviewResponse = `<body>
<h1>Overview</h1>
<h2>Time series</h2>
<a href="/images/week.png"><img src="/images/week.png" alt="time series" style="width:400px;height:400px;"></a>
<h2>Summary</h2>
<a href="/images/summary.png"><img src="/images/summary.png" alt="time series" style="width:400px;height:400px;"></a>
</body>`

func (server *Server) overview(w http.ResponseWriter, req *http.Request) {
	start, stop, err := server.parseGenerateRequest(req)

	if err != nil {
		http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	err = server.backend.Overview(start, stop)

	if err != nil {
		http.Error(w, "unable to create image: "+err.Error(), http.StatusInternalServerError)
		return
	}

	t := template.New("body")
	t, err = t.Parse(OverviewResponse)

	if err != nil {
		http.Error(w, "unable to display page: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_ = t.Execute(w, nil)
	// w.WriteHeader(http.StatusOK)
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
}

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

	if ok == false {
		return dbfunc()
	}

	timestamp, err = time.Parse(time.RFC3339, arg[0])

	if err != nil {
		err = fmt.Errorf("invalid format for '%s' argument: %s", field, err.Error())
	}
	return
}
