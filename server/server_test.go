package server_test

import (
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"path"
	"testing"
)

func TestServer_Overview_Arguments(t *testing.T) {
	tmpdir, _ := os.MkdirTemp("", "")
	s := server.New(8081, reports.New(tmpdir, mockdb.BuildDB()))
	go s.Run()

	testCases := []struct {
		url          string
		responseCode int
	}{
		{url: "http://localhost:8081/overview?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK},
		{url: "http://localhost:8081/overview", responseCode: http.StatusOK},
		{url: "http://localhost:8081/overview?start=123&stop=123", responseCode: http.StatusBadRequest},
		{url: "http://localhost:8081/overview?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
	}

	for _, testCase := range testCases {
		resp, err := http.Get(testCase.url)
		if assert.NoError(t, err, testCase.url) {
			assert.Equal(t, testCase.responseCode, resp.StatusCode, testCase.responseCode)
			_ = resp.Body.Close()
		}
	}

	_ = os.RemoveAll(tmpdir)
}

func TestServer_Overview(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)
	s := server.New(8082, reports.New(tmpdir, mockdb.BuildDB()))
	go s.Run()

	resp, err := http.Get("http://localhost:8082/overview?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z")

	if assert.NoError(t, err) {
		_ = resp.Body.Close()
		if assert.Equal(t, http.StatusOK, resp.StatusCode) {
			for _, filename := range []string{"summary.png", "week.png"} {
				assert.FileExists(t, path.Join(tmpdir, filename), filename)

				resp, err = http.Get("http://localhost:8081/images/" + filename)
				if assert.NoError(t, err, filename) {
					// assert.Equal(t, "image/png", resp.Header["Content-Type"][0])
					_ = resp.Body.Close()
				}
			}
		}
	}

	err = os.RemoveAll(tmpdir)
	assert.NoError(t, err)
}
