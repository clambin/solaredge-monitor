package web

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_parseTimestamp(t *testing.T) {
	location := time.FixedZone("", 3600)
	testCases := []struct {
		name string
		arg  string
		want time.Time
		err  assert.ErrorAssertionFunc
	}{
		{
			name: "2006-01-02T15:04:05Z07:00 (RFC3339)",
			arg:  `2023-08-25T00:00:00+01:00`,
			want: time.Date(2023, time.August, 25, 0, 0, 0, 0, location),
			err:  assert.NoError,
		},
		{
			name: "2006-01-02T15:04:05",
			arg:  `2023-08-25T01:00:00`,
			want: time.Date(2023, time.August, 25, 1, 0, 0, 0, time.UTC),
			err:  assert.NoError,
		},
		{
			name: "2006-01-02T15:04",
			arg:  `2023-08-25T01:00`,
			want: time.Date(2023, time.August, 25, 1, 0, 0, 0, time.UTC),
			err:  assert.NoError,
		},
		{
			name: "2006-01-02",
			arg:  `2023-08-25`,
			want: time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC),
			err:  assert.NoError,
		},
		{
			name: "invalid date",
			arg:  `invalid-date`,
			err:  assert.Error,
		},
		{
			name: "blank date",
			arg:  ``,
			want: time.Time{},
			err:  assert.NoError,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			args, err := parseTimestamp(tt.arg)
			tt.err(t, err)
			assert.Equal(t, tt.want, args)
		})
	}
}
