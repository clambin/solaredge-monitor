package configuration

import (
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
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
	if content, err = os.ReadFile(filename); err == nil {
		config, err = Load(content)
	}
	return
}

func Load(content []byte) (configuration *Configuration, err error) {
	configuration = &Configuration{
		Server: ServerConfiguration{
			Port:   80,
			Images: "/images",
		},
		Scrape: ScrapeConfiguration{
			Polling:    5 * time.Minute,
			Collection: 15 * time.Minute,
		},
		Database: loadDBEnvironment(),
	}
	err = yaml.Unmarshal(content, &configuration)
	return
}

func loadDBEnvironment() DBConfiguration {
	var (
		err        error
		pgHost     string
		pgPort     int
		pgDatabase string
		pgUser     string
		pgPassword string
	)
	if pgHost = os.Getenv("pg_host"); pgHost == "" {
		pgHost = "postgres"
	}
	if pgPort, err = strconv.Atoi(os.Getenv("pg_port")); err != nil || pgPort == 0 {
		pgPort = 5432
	}
	if pgDatabase = os.Getenv("pg_database"); pgDatabase == "" {
		pgDatabase = "solar"
	}
	if pgUser = os.Getenv("pg_user"); pgUser == "" {
		pgUser = "solar"
	}
	if pgPassword = os.Getenv("pg_password"); pgPassword == "" {
		pgPassword = "solar"
	}

	return DBConfiguration{
		Host:     pgHost,
		Port:     pgPort,
		Database: pgDatabase,
		Username: pgUser,
		Password: pgPassword,
	}
}
