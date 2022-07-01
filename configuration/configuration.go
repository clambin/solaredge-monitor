package configuration

import (
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type Configuration struct {
	Debug     bool                   `yaml:"debug"`
	Server    ServerConfiguration    `yaml:"server"`
	Scrape    ScrapeConfiguration    `yaml:"scrape"`
	Database  DBConfiguration        `yaml:"database"`
	Tado      TadoConfiguration      `yaml:"tado"`
	SolarEdge SolarEdgeConfiguration `yaml:"solarEdge"`
}

type ServerConfiguration struct {
	Port   int    `yaml:"port"`
	Images string `yaml:"images"`
}

type ScrapeConfiguration struct {
	Enabled    bool          `yaml:"enabled"`
	Polling    time.Duration `yaml:"polling"`
	Collection time.Duration `yaml:"collection"`
}

type DBConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type TadoConfiguration struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type SolarEdgeConfiguration struct {
	Token string `yaml:"token"`
}

func LoadFromFile(filename string) (config *Configuration, err error) {
	var content []byte
	if content, err = os.ReadFile(filename); err != nil {
		return
	}
	content = []byte(os.ExpandEnv(string(content)))

	config = &Configuration{
		Server: ServerConfiguration{
			Port:   80,
			Images: "/images",
		},
		Scrape: ScrapeConfiguration{
			Polling:    5 * time.Minute,
			Collection: 15 * time.Minute,
		},
		Database: DBConfiguration{
			Host:     "postgres",
			Port:     5432,
			Database: "solar",
			Username: "solar",
			Password: "solar",
		},
	}
	err = yaml.Unmarshal(content, config)
	return
}
