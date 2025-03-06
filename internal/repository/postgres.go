package repository

import (
	"embed"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"net/url"
	"strings"
	"time"
)

var _ prometheus.Collector = &PostgresDB{}

type PostgresDB struct {
	prometheus.Collector
	DBX      *sqlx.DB
	database string
}

func NewPostgresDB(connectionString string) (*PostgresDB, error) {
	dbName, err := getDBName(connectionString)
	if err != nil {
		return nil, fmt.Errorf("invalid db url %q: %w", connectionString, err)
	}
	var db *PostgresDB
	dbx, err := sqlx.Connect("postgres", connectionString)
	if err == nil {
		db = &PostgresDB{
			database:  dbName,
			DBX:       dbx,
			Collector: collectors.NewDBStatsCollector(dbx.DB, dbName),
		}
		err = db.migrate()
	}
	return db, err
}

func getDBName(connectionString string) (string, error) {
	u, err := url.Parse(connectionString)
	if err != nil {
		return "", err
	}
	if u.Scheme != "postgres" {
		return "", errors.New("not a postgres url")
	}
	if u.Path == "" || u.Path == "/" {
		return "", errors.New("no database specified")
	}
	return u.Path[1:], nil
}

func (db *PostgresDB) Store(measurement Measurement) error {
	weatherID, err := db.GetWeatherID(measurement.Weather)
	if err == nil {
		_, err = db.DBX.Exec(`INSERT INTO solar (timestamp, intensity, power, weatherid) VALUES ($1, $2, $3, $4)`,
			measurement.Timestamp, measurement.Intensity, measurement.Power, weatherID,
		)
	}
	return err
}

func (db *PostgresDB) Get(from, to time.Time) (Measurements, error) {
	stmt := "SELECT timestamp, intensity, power, weather FROM solar, weatherids WHERE solar.weatherid = weatherids.id"
	if timeClause := getTimeClause(from, to); timeClause != "" {
		stmt += " AND " + timeClause
	}
	stmt += " ORDER BY timestamp"
	var measurements Measurements
	err := db.DBX.Select(&measurements, stmt)
	return measurements, err
}

func getTimeClause(from, to time.Time) string {
	conditions := make([]string, 0, 2)
	if !from.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= '%s'", from.Format("2006-01-02 15:04:05")))
	}
	if !to.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= '%s'", to.Format("2006-01-02 15:04:05")))
	}
	return strings.Join(conditions, " AND ")
}

func (db *PostgresDB) GetDataRange() (time.Time, time.Time, error) {
	var response struct {
		First time.Time `db:"first"`
		Last  time.Time `db:"last"`
	}
	err := db.DBX.Get(&response, `SELECT MIN(timestamp) "first", MAX(timestamp) "last" FROM solar`)
	return response.First, response.Last, err
}

//go:embed migrations/*
var migrations embed.FS

func (db *PostgresDB) migrate() error {
	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("invalid migration source: %w", err)
	}

	dbDriver, err := postgres.WithInstance(db.DBX.DB, &postgres.Config{DatabaseName: db.database})
	if err != nil {
		return fmt.Errorf("invalid migration target: %w", err)
	}

	m, err := migrate.NewWithInstance("migrations", src, db.database, dbDriver)
	if err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("database migration failed: %w", err)
	}

	return nil
}
