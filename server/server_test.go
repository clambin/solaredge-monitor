package server_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/clambin/solaredge-monitor/server"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestServer_Report(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(8081, mockdb.BuildDB())
	go s.Run(ctx)

	require.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:8081/")
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil
	}, 5*time.Second, 100*time.Millisecond)

	testCases := []struct {
		args         string
		responseCode int
		searchString string
	}{
		{args: "start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK, searchString: "<title>Report</title>"},
		{args: "", responseCode: http.StatusOK, searchString: "<title>Report</title>"},
		{args: "start=123&stop=123", responseCode: http.StatusBadRequest},
		{args: "start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
	}

	for index, testCase := range testCases {
		url := "http://127.0.0.1:8081/report"
		if testCase.args != "" {
			url += "?" + testCase.args
		}

		resp, err := http.Get(url)
		require.NoError(t, err, testCase.args)
		assert.Equal(t, testCase.responseCode, resp.StatusCode, testCase.responseCode)

		if testCase.responseCode != http.StatusOK {
			continue
		}

		var buffer, golden []byte
		buffer, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		gp := fmt.Sprintf("testdata/%s_%d.golden", t.Name(), index)
		if *update {
			err = os.WriteFile(gp, buffer, 0644)
			require.NoError(t, err)
		}

		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(buffer, golden))

		_ = resp.Body.Close()
	}

	cancel()
}

func TestServer_Plot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := server.New(8082, mockdb.BuildDB())
	go s.Run(ctx)

	require.Eventually(t, func() bool {
		resp, err := http.Get("http://127.0.0.1:8082/")
		if err == nil {
			_ = resp.Body.Close()
		}
		return err == nil
	}, 5*time.Second, 100*time.Millisecond)

	testCases := []struct {
		args         string
		responseCode int
	}{
		{args: "type=scatter", responseCode: http.StatusOK},
		// TODO: contour output not always the same (even though it is in plotter)?
		// {args: "type=contour", responseCode: http.StatusOK},
		{args: "type=heatmap", responseCode: http.StatusOK},
		{args: "start=2020-06-25T21:19:00.000Z&stop=2021-06-25T21:19:00.000Z", responseCode: http.StatusOK},
		{args: "", responseCode: http.StatusOK},
		{args: "type=notatype", responseCode: http.StatusBadRequest},
		{args: "start=123&stop=123", responseCode: http.StatusBadRequest},
		{args: "start=2021-06-25T21:19:00.000Z&stop=2020-06-25T21:19:00.000Z", responseCode: http.StatusBadRequest},
	}

	for index, testCase := range testCases {
		url := "http://127.0.0.1:8082/plot"
		if testCase.args != "" {
			url += "?" + testCase.args
		}

		resp, err := http.Get(url)
		require.NoError(t, err, testCase.args)
		assert.Equal(t, testCase.responseCode, resp.StatusCode, index)

		if testCase.responseCode != http.StatusOK {
			continue
		}

		var buffer, golden []byte
		buffer, err = io.ReadAll(resp.Body)
		require.NoError(t, err)

		gp := fmt.Sprintf("testdata/%s_%d.golden", t.Name(), index)
		if *update {
			err = os.WriteFile(gp, buffer, 0644)
			require.NoError(t, err)
		}

		golden, err = os.ReadFile(gp)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(buffer, golden), index)
		//assert.Equal(t, golden, buffer, index)

		_ = resp.Body.Close()
	}

	cancel()
}
