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
	"time"
)

func TestServer_Overview(t *testing.T) {
	tmpdir, _ := os.MkdirTemp("", "")
	s := server.New(8081, reports.New(tmpdir, mockdb.BuildDB()))
	go s.Run()

	var resp *http.Response
	var err error
	if assert.Eventually(t, func() bool {
		resp, err = http.Get("http://localhost:8081/overview")
		return err == nil
	}, 500*time.Millisecond, 50*time.Millisecond) {

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
			resp, err = http.Get(testCase.url)
			if assert.NoError(t, err, testCase.url) {
				assert.Equal(t, testCase.responseCode, resp.StatusCode, testCase.responseCode)
				_ = resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					for _, filename := range []string{"summary.png", "week.png"} {
						assert.FileExists(t, path.Join(tmpdir, filename), filename)

						resp, err = http.Get("http://localhost:8081/images/" + filename)
						if assert.NoError(t, err, filename) {
							assert.Equal(t, "image/png", resp.Header["Content-Type"][0])
							_ = resp.Body.Close()
						}
					}
				}
			}
		}
	}

	_ = os.RemoveAll(tmpdir)
}
