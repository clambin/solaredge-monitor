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

func parseTimestamp(arg string) (timestamp time.Time, err error) {
	if arg == "" {
		return
	}
	if timestamp, err = time.Parse(time.RFC3339, arg); err != nil {
		err = fmt.Errorf("invalid format for %q: %w", arg, err)
	}
	return timestamp, err
}
