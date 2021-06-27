package server_test

import (
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	server2 "github.com/clambin/solaredge-monitor/web/server"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

func TestServer_Overview(t *testing.T) {
	s := server2.New(8081, "", reports.New(mockdb.BuildDB()))
	go s.Run()

	var resp *http.Response
	var err error
	if assert.Eventually(t, func() bool {
		resp, err = http.Get("http://localhost:8081/")
		return err == nil
	}, 5*time.Second, 100*time.Millisecond) {

		testCases := []struct {
			url          string
			responseCode int
		}{
			{url: "http://localhost:8081/report?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK},
			{url: "http://localhost:8081/report", responseCode: http.StatusOK},
			{url: "http://localhost:8081/report?start=123&stop=123", responseCode: http.StatusBadRequest},
			{url: "http://localhost:8081/report?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
		}

		for _, testCase := range testCases {
			resp, err = http.Get(testCase.url)
			if assert.NoError(t, err, testCase.url) {
				assert.Equal(t, testCase.responseCode, resp.StatusCode, testCase.responseCode)
				_ = resp.Body.Close()
			}
		}
	}
}
