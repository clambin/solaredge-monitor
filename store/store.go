package store

import (
	"embed"
	"fmt"
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
	psqlInfo string
	database string
	DBH      *sqlx.DB
}

func NewPostgresDB(host string, port int, database string, user string, password string) (db *PostgresDB, err error) {
	db = &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		database: database,
	}
	err = db.initialize()
	return db, err
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

func (db *PostgresDB) initialize() (err error) {
	if db.DBH, err = sqlx.Connect("postgres", db.psqlInfo); err != nil {
		return fmt.Errorf("database connect: %w", err)
	}

	if err = db.migrate(); err != nil {
		return fmt.Errorf("database migration: %w", err)
	}

	prometheus.DefaultRegisterer.MustRegister(collectors.NewDBStatsCollector(db.DBH.DB, db.database))

	return nil
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
