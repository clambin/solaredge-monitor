package server

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestServer_Handlers(t *testing.T) {
	s := New(mockdb.BuildDB())

	testCases := []struct {
		path         string
		responseCode int
	}{
		{path: "/plot/scatter", responseCode: http.StatusOK},
		// TODO: contour output not always the same (even though it is in plotter)?
		// {args: "type=contour", responseCode: http.StatusOK},
		{path: "/plot/heatmap", responseCode: http.StatusOK},
		{path: "/plot/scatter?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK},
		{path: "/plot/notatype", responseCode: http.StatusBadRequest},
		{path: "/plot/scatter?start=123&stop=123", responseCode: http.StatusBadRequest},
		{path: "/plot/scatter?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
		{path: "/report", responseCode: http.StatusOK},
		{path: "/report?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK},
		{path: "/report?start=123&stop=123", responseCode: http.StatusBadRequest},
		{path: "/report?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
		{path: "/", responseCode: http.StatusSeeOther},
	}

	for index, testCase := range testCases {
		url := "http://127.0.0.1" + testCase.path
		w := httptest.NewRecorder()
		r, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		s.ServeHTTP(w, r)
		assert.Equal(t, testCase.responseCode, w.Result().StatusCode, url)

		if testCase.responseCode == http.StatusSeeOther {
			assert.Equal(t, "/report", w.Header().Get("Location"))
		}

		if testCase.responseCode != http.StatusOK {
			continue
		}

		var buffer, golden []byte
		buffer, err = io.ReadAll(w.Body)
		require.NoError(t, err)

		gp := filepath.Join("testdata", fmt.Sprintf("%s_%d.golden", strings.ToLower(t.Name()), index))
		if *update {
			err = os.WriteFile(gp, buffer, 0644)
			require.NoError(t, err)
		}

		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(buffer, golden), index)
	}
}

func TestServer_ParseTimestamps(t *testing.T) {
	s := New(mockdb.BuildDB())

	validFromString := "2006-01-02T15:04:05Z"
	validFrom, err := time.Parse(time.RFC3339, validFromString)
	require.NoError(t, err)
	validToString := "2006-01-03T15:04:05Z"
	validTo, err := time.Parse(time.RFC3339, validToString)
	require.NoError(t, err)

	var tests = []struct {
		name         string
		from         string
		to           string
		pass         bool
		expectedFrom time.Time
		expectedTo   time.Time
	}{
		{name: "empty", pass: true},
		{name: "valid from", from: validFromString, pass: true, expectedFrom: validFrom},
		{name: "valid to", to: validToString, pass: true, expectedTo: validTo},
		{name: "valid from and to", from: validFromString, to: validToString, pass: true, expectedFrom: validFrom, expectedTo: validTo},
		{name: "invalid from", from: "foo", pass: false},
		{name: "invalid to", to: "foo", pass: false},
		{name: "invalid from and to", from: validToString, to: validFromString, pass: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			q := req.URL.Query()
			if tt.from != "" {
				q.Add("start", tt.from)
			}
			if tt.to != "" {
				q.Add("stop", tt.to)
			}
			req.URL.RawQuery = q.Encode()

			from, to, err := s.parseTimestamps(req)

			if !tt.pass {
				assert.Error(t, err)
				return
			}

			if !tt.expectedFrom.IsZero() {
				assert.Equal(t, tt.expectedFrom, from)
			}
			if !tt.expectedTo.IsZero() {
				assert.Equal(t, tt.expectedTo, to)
			}
		})
	}
}
