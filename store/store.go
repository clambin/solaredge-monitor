package store

import (
	"embed"
	"fmt"
	"github.com/clambin/solaredge-monitor/configuration"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"time"
)

type DB interface {
	Store(Measurement) error
	Get(time.Time, time.Time) ([]Measurement, error)
	GetAll() ([]Measurement, error)
}

type Measurement struct {
	Timestamp time.Time
	Power     float64
	Intensity float64
}

type PostgresDB struct {
	database  string
	DBH       *sqlx.DB
	collector prometheus.Collector
}

var _ DB = &PostgresDB{}

func NewPostgresDBFromConfig(cfg configuration.DBConfiguration) (db *PostgresDB, err error) {
	return NewPostgresDB(
		cfg.Host, cfg.Port,
		cfg.Database,
		cfg.Username, cfg.Password,
	)
}

func NewPostgresDB(host string, port int, database string, user string, password string) (*PostgresDB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, database)
	dbh, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("database connect: %w", err)
	}

	db := &PostgresDB{
		database:  database,
		DBH:       dbh,
		collector: collectors.NewDBStatsCollector(dbh.DB, database),
	}

	if err = db.migrate(); err != nil {
		return nil, fmt.Errorf("database migration: %w", err)
	}

	return db, err
}

func (db *PostgresDB) Describe(descs chan<- *prometheus.Desc) {
	db.collector.Describe(descs)
}

func (db *PostgresDB) Collect(metrics chan<- prometheus.Metric) {
	db.collector.Collect(metrics)
}

func (db *PostgresDB) Store(measurement Measurement) (err error) {
	tx := db.DBH.MustBegin()
	tx.MustExec(`INSERT INTO solar(timestamp, intensity, power) VALUES ($1, $2, $3)`,
		measurement.Timestamp, measurement.Intensity, measurement.Power,
	)
	return tx.Commit()
}

func (db *PostgresDB) Get(from, to time.Time) (measurements []Measurement, err error) {
	err = db.DBH.Select(&measurements, fmt.Sprintf(`SELECT timestamp, intensity, power FROM solar WHERE timestamp >= '%s' AND timestamp <= '%s' ORDER BY 1`,
		from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05"),
	))
	return
}

func (db *PostgresDB) GetAll() (measurements []Measurement, err error) {
	err = db.DBH.Select(&measurements, `SELECT timestamp, intensity, power FROM solar ORDER BY 1`)
	return
}

//go:embed migrations/*
var migrations embed.FS

func (db *PostgresDB) migrate() error {
	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("invalid migration source: %w", err)
	}

	dbDriver, err := postgres.WithInstance(db.DBH.DB, &postgres.Config{DatabaseName: db.database})
	if err != nil {
		return fmt.Errorf("invalid migration target: %w", err)
	}

	m, err := migrate.NewWithInstance("migrations", src, db.database, dbDriver)
	if err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	if err = m.Up(); err == migrate.ErrNoChange {
		err = nil
	}

	return err
}
