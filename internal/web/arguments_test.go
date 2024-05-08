package web_test

import (
	"github.com/clambin/solaredge-monitor/internal/web"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	location := time.FixedZone("UK", int(time.Hour.Seconds()))

	testCases := []struct {
		name    string
		args    url.Values
		want    web.Arguments
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "default",
			args:    url.Values{},
			wantErr: assert.NoError,
		},
		{
			name:    "start",
			args:    url.Values{"start": []string{`2023-08-25T00:00:00+00:00`}},
			want:    web.Arguments{Start: time.Date(2023, time.August, 25, 0, 0, 0, 0, time.UTC)},
			wantErr: assert.NoError,
		},
		{
			name:    "stop",
			args:    url.Values{"stop": []string{`2023-08-25T00:00:00+01:00`}},
			want:    web.Arguments{Stop: time.Date(2023, time.August, 25, 0, 0, 0, 0, location)},
			wantErr: assert.NoError,
		},
		{
			name:    "fold",
			args:    url.Values{"fold": []string{"true"}},
			want:    web.Arguments{Fold: true},
			wantErr: assert.NoError,
		},
		{
			name:    "invalid date",
			args:    url.Values{"stop": []string{`invalid-date`}},
			wantErr: assert.Error,
		},
		{
			name:    "swapped dates",
			args:    url.Values{"start": []string{`2023-08-25T12:00:00+01:00`}, "stop": []string{`2023-08-25T00:00:00+01:00`}},
			wantErr: assert.Error,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/?"+tt.args.Encode(), nil)

			args, err := web.Parse(req)
			tt.wantErr(t, err)

			if err == nil {
				assert.True(t, tt.want.Start.Equal(args.Start))
				assert.True(t, tt.want.Stop.Equal(args.Stop))
				assert.Equal(t, tt.want.Fold, args.Fold)
			}
		})
	}
}
