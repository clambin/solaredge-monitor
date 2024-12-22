package web

import (
	"errors"
	"time"
)

var timeFormats = []string{
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02",
}

func parseTimestamp(arg string) (time.Time, error) {
	if arg == "" {
		return time.Time{}, nil
	}
	for _, timeFormat := range timeFormats {
		if timestamp, err := time.Parse(timeFormat, arg); err == nil {
			return timestamp, nil
		}
	}
	return time.Time{}, errors.New("invalid timestamp")
}
