package server_test

import (
	"context"
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/clambin/solaredge-monitor/web/server"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestServer_Overview(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(8081, "", reports.New(mockdb.BuildDB()))
	go s.Run(ctx)

	var resp *http.Response
	var err error
	if assert.Eventually(t, func() bool {
		resp, err = http.Get("http://127.0.0.1:8081/")
		return err == nil
	}, 5*time.Second, 100*time.Millisecond) {

		testCases := []struct {
			url          string
			responseCode int
			searchString string
		}{
			{url: "http://127.0.0.1:8081/report?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK, searchString: "<title>Report</title>"},
			{url: "http://127.0.0.1:8081/report", responseCode: http.StatusOK, searchString: "<title>Report</title>"},
			{url: "http://127.0.0.1:8081/report?start=123&stop=123", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/report?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/summary?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK, searchString: "<title>Summary</title>"},
			{url: "http://127.0.0.1:8081/summary", responseCode: http.StatusOK, searchString: "<title>Summary</title>"},
			{url: "http://127.0.0.1:8081/summary?start=123&stop=123", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/summary?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/timeseries?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK, searchString: "<title>Time Series</title>"},
			{url: "http://127.0.0.1:8081/timeseries", responseCode: http.StatusOK, searchString: "<title>Time Series</title>"},
			{url: "http://127.0.0.1:8081/timeseries?start=123&stop=123", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/timeseries?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/classify?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK, searchString: "<title>Classification</title>"},
			{url: "http://127.0.0.1:8081/classify", responseCode: http.StatusOK, searchString: "<title>Classification</title>"},
			{url: "http://127.0.0.1:8081/classify?start=123&stop=123", responseCode: http.StatusBadRequest},
			{url: "http://127.0.0.1:8081/classify?start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
		}

		for _, testCase := range testCases {
			resp, err = http.Get(testCase.url)
			if assert.NoError(t, err, testCase.url) {
				assert.Equal(t, testCase.responseCode, resp.StatusCode, testCase.responseCode)
				if testCase.responseCode == http.StatusOK && testCase.searchString != "" {
					var buffer []byte
					buffer, err = io.ReadAll(resp.Body)

					assert.NoError(t, err)
					assert.Contains(t, string(buffer), testCase.searchString)
				}
				_ = resp.Body.Close()
			}
		}
	}

	cancel()
}

func TestServer_BadDirectory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(8082, "/notadirectory", reports.New(mockdb.BuildDB()))
	go s.Run(ctx)

	var err error
	if assert.Eventually(t, func() bool {
		_, err = http.Get("http://127.0.0.1:8082/")
		return err == nil
	}, 5*time.Second, 100*time.Millisecond) {

		for _, url := range []string{
			"http://127.0.0.1:8082/report?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z",
			"http://127.0.0.1:8082/summary?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z",
			"http://127.0.0.1:8082/timeseries?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z",
			"http://127.0.0.1:8082/classify?start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z",
		} {
			var resp *http.Response
			resp, err = http.Get(url)

			assert.NoError(t, err, url)
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, url)
		}

	}

	cancel()
}
