package handlers

import (
	"fmt"
	"net/http"
	"time"
)

type arguments struct {
	start time.Time
	stop  time.Time
	fold  bool
}

func parseArguments(req *http.Request) (args arguments, err error) {
	fold := req.URL.Query().Get("fold")
	if fold == "true" {
		args.fold = true
	}

	if args.start, err = parseTimestamp(req, "start"); err != nil {
		return arguments{}, err
	}
	if args.stop, err = parseTimestamp(req, "stop"); err != nil {
		return
	}

	if !args.stop.IsZero() && args.stop.Before(args.start) {
		err = fmt.Errorf("start time is later than stop time")
	}

	return
}

func parseTimestamp(req *http.Request, field string) (timestamp time.Time, err error) {
	arg := req.URL.Query().Get(field)
	if arg == "" {
		return
	}
	if timestamp, err = time.Parse(time.RFC3339, arg); err != nil {
		err = fmt.Errorf("invalid format for '%s': %w", field, err)
	}
	return timestamp, err
}
