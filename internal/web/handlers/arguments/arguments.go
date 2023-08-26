package arguments

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
	fold := req.URL.Query().Get("fold")
	if fold == "true" {
		args.Fold = true
	}

	if args.Start, err = parseTimestamp(req, "start"); err != nil {
		return Arguments{}, err
	}
	if args.Stop, err = parseTimestamp(req, "stop"); err != nil {
		return
	}

	if !args.Stop.IsZero() && args.Stop.Before(args.Start) {
		err = fmt.Errorf("start time is later than Stop time")
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
