package web

import (
	"fmt"
	"net/http"
	"time"
)

type Arguments struct {
	Start time.Time
	Stop  time.Time
	Fold  bool
}

func Parse(req *http.Request) (args Arguments, err error) {
	values := req.URL.Query()
	fold := values.Get("fold")
	if fold == "true" {
		args.Fold = true
	}

	if args.Start, err = parseTimestamp(values.Get("start")); err != nil {
		return Arguments{}, err
	}
	if args.Stop, err = parseTimestamp(values.Get("stop")); err != nil {
		return
	}

	if !args.Stop.IsZero() && args.Stop.Before(args.Start) {
		err = fmt.Errorf("start time is later than Stop time")
	}

	return
}

func parseTimestamp(arg string) (timestamp time.Time, err error) {
	if arg == "" {
		return
	}
	if timestamp, err = time.Parse(time.RFC3339, arg); err != nil {
		err = fmt.Errorf("invalid format for '%s': %w", arg, err)
	}
	return timestamp, err
}
