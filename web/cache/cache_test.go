package cache_test

import (
	"github.com/clambin/solaredge-monitor/web/cache"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
	"time"
)

func TestCache_Run(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	tmpdir, _ := os.MkdirTemp("", "")

	c := cache.New(tmpdir, 250*time.Millisecond, 25*time.Millisecond)
	go c.Run()

	filename, err := c.Store("test", []byte(``))
	assert.NoError(t, err)
	assert.NotEqual(t, "test", filename)

	assert.Eventually(t, func() bool {
		_, err = os.Stat(path.Join(tmpdir, filename))
		return os.IsNotExist(err)
	}, 500*time.Millisecond, 50*time.Millisecond)
}
