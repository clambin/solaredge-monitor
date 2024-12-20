package web

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
	values := req.URL.Query()
	fold := values.Get("fold")
	if fold == "true" {
		args.fold = true
	}

	if args.start, err = parseTimestamp(values.Get("start")); err != nil {
		return arguments{}, err
	}
	if args.stop, err = parseTimestamp(values.Get("stop")); err != nil {
		return
	}

	if !args.stop.IsZero() && args.stop.Before(args.start) {
		err = fmt.Errorf("start time is later than Stop time")
	}

	return
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04",
}

func parseTimestamp(arg string) (timestamp time.Time, err error) {
	if arg == "" {
		return
	}
	for _, timeFormat := range timeFormats {
		if timestamp, err = time.Parse(timeFormat, arg); err == nil {
			return timestamp, nil
		}
	}
	return timestamp, fmt.Errorf("invalid format for %q: %w", arg, err)
}
