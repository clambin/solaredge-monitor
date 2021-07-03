package cache

import (
	"context"
	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

type Cache struct {
	Add       chan string
	Directory string

	refresh   *time.Ticker
	retention time.Duration
	content   map[string]time.Time
}

func New(directory string, retention time.Duration, interval time.Duration) *Cache {
	return &Cache{
		Add:       make(chan string),
		Directory: directory,
		refresh:   time.NewTicker(interval),
		retention: retention,
		content:   make(map[string]time.Time),
	}
}

func (cache *Cache) Run(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case filename := <-cache.Add:
			cache.add(filename)
		case <-cache.refresh.C:
			cache.cleanup()
		}
	}
}

func (cache *Cache) add(filename string) {
	cache.content[filename] = time.Now().Add(cache.retention)

	log.WithFields(log.Fields{
		"filename": filename,
		"expiry":   cache.content[filename],
	}).Debug("file added to file cache")
}

func (cache *Cache) cleanup() {
	expired := make([]string, 0)
	for filename, expiry := range cache.content {
		if time.Now().After(expiry) {
			log.WithField("filename", filename).Debug("removing expired file from cache")

			if err := os.Remove(path.Join(cache.Directory, filename)); err != nil {
				log.WithError(err).WithField("filename", filename).Warning("failed to remove cached file")
			}
			expired = append(expired, filename)
		}
	}

	for _, filename := range expired {
		delete(cache.content, filename)
	}
}

func (cache *Cache) Store(filename string, content []byte) (unique string, err error) {
	unique = uuid.NewV4().String() + "-" + filename

	err = os.WriteFile(path.Join(cache.Directory, unique), content, 0600)

	if err == nil {
		cache.Add <- unique
	} else {
		log.WithError(err).Error("failed to store file in cache")
	}

	return
}
