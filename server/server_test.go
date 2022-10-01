package server

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var update = flag.Bool("update", false, "update .golden files")

func TestServer_Handlers(t *testing.T) {
	s := New(0, 0, mockdb.BuildDB())

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
		rr := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err)

		s.httpServers["app"].Handler.ServeHTTP(rr, req)
		assert.Equal(t, testCase.responseCode, rr.Result().StatusCode, url)

		if testCase.responseCode == http.StatusSeeOther {
			assert.Equal(t, "/report", rr.Header().Get("Location"))
		}

		if testCase.responseCode != http.StatusOK {
			continue
		}

		var buffer, golden []byte
		buffer, err = io.ReadAll(rr.Body)
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

func TestServer_Run(t *testing.T) {
	s := New(8081, 9092, mockdb.BuildDB())

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		s.Run(ctx)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		resp, err := http.Get("http://localhost:9092/metrics")
		if err != nil {
			return false
		}
		_ = resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, time.Second, 10*time.Millisecond)

	cancel()
	wg.Wait()
}

func TestServer_Run_Error(t *testing.T) {
	var exitCode int
	var lock sync.Mutex

	logger := log.StandardLogger()
	logger.ExitFunc = func(code int) {
		lock.Lock()
		defer lock.Unlock()
		exitCode = code
	}

	s := New(8081, 8081, mockdb.BuildDB())

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.Run(ctx)
		wg.Done()
	}()

	assert.Eventually(t, func() bool {
		lock.Lock()
		defer lock.Unlock()
		return exitCode != 0
	}, time.Second, 10*time.Millisecond)

	cancel()
	wg.Wait()
}
