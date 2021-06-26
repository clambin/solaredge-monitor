package configuration_test

import (
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestLoadFromFile(t *testing.T) {
	const config = `
debug: true
server:
  port: 8080
  images: /static
scrape:
  enabled: true
  polling: 15m
  collection: 1h
database:
  host: localhost
  port: 1234
  database: test
  username: user
  password: password
tado:
  username: tado
  password: tadopassword
solarEdge:
  token: 1234
`

	var (
		err error
		f   *os.File
		cfg *configuration.Configuration
	)

	f, err = os.CreateTemp("", "tmp")
	if assert.NoError(t, err) {
		defer func(name string) {
			_ = os.Remove(name)
		}(f.Name())

		_, _ = f.Write([]byte(config))
		_ = f.Close()

		cfg, err = configuration.LoadFromFile(f.Name())

		assert.NoError(t, err)
		assert.True(t, cfg.Debug)
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "/static", cfg.Server.Images)
		assert.True(t, cfg.Scrape.Enabled)
		assert.Equal(t, 15*time.Minute, cfg.Scrape.Polling)
		assert.Equal(t, 1*time.Hour, cfg.Scrape.Collection)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, 1234, cfg.Database.Port)
		assert.Equal(t, "test", cfg.Database.Database)
		assert.Equal(t, "user", cfg.Database.Username)
		assert.Equal(t, "password", cfg.Database.Password)
		assert.Equal(t, "tado", cfg.Tado.Username)
		assert.Equal(t, "tadopassword", cfg.Tado.Password)
		assert.Equal(t, "1234", cfg.SolarEdge.Token)
	}
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := configuration.Load([]byte(``))

	assert.NoError(t, err)
	assert.Equal(t, 80, cfg.Server.Port)
	assert.Equal(t, "/images", cfg.Server.Images)
	assert.False(t, cfg.Scrape.Enabled)
	assert.Equal(t, 5*time.Minute, cfg.Scrape.Polling)
	assert.Equal(t, 15*time.Minute, cfg.Scrape.Collection)
	assert.Equal(t, "postgres", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "solar", cfg.Database.Database)
	assert.Equal(t, "solar", cfg.Database.Username)
	assert.Equal(t, "solar", cfg.Database.Password)
}
