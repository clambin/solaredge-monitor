package reports_test

import (
	"github.com/clambin/solaredge-monitor/reports"
	"github.com/clambin/solaredge-monitor/store/mockdb"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestServer_Overview(t *testing.T) {
	db := mockdb.BuildDB()

	tmpdir, _ := os.MkdirTemp("", "overview-")
	reporter := reports.New(tmpdir, db)

	start, _ := reporter.GetFirst()
	stop, _ := reporter.GetLast()

	assert.NotEqual(t, start, stop)

	err := reporter.Overview(start, stop)
	assert.NoError(t, err)

	assert.FileExists(t, path.Join(tmpdir, "summary.png"))
	assert.FileExists(t, path.Join(tmpdir, "week.png"))

	_ = os.RemoveAll(tmpdir)
}
