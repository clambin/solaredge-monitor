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
	"strings"
	"time"
)

type PostgresDB struct {
	prometheus.Collector
	database string
	DBH      *sqlx.DB
}

var _ prometheus.Collector = &PostgresDB{}

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
		Collector: collectors.NewDBStatsCollector(dbh.DB, database),
	}

	err = db.migrate()
	return db, err
}

func (db *PostgresDB) Store(measurement Measurement) error {
	weatherID, err := db.GetWeatherID(measurement.Weather)
	if err == nil {
		tx := db.DBH.MustBegin()
		tx.MustExec(`INSERT INTO solar(timestamp, intensity, power, weatherid) VALUES ($1, $2, $3, $4)`,
			measurement.Timestamp, measurement.Intensity, measurement.Power, weatherID,
		)
		err = tx.Commit()
	}
	return err
}

func (db *PostgresDB) Get(from, to time.Time) (measurements Measurements, err error) {
	err = db.DBH.Select(&measurements,
		fmt.Sprintf(`SELECT timestamp, intensity, power, weather FROM solar, weatherids WHERE solar.weatherid = weatherids.id %s ORDER BY 1`,
			getTimeClause(from, to),
		),
	)
	return
}

func getTimeClause(from, to time.Time) string {
	var conditions []string
	if !from.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= '%s'", from.Format("2006-01-02 15:04:05")))
	}
	if !to.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= '%s'", to.Format("2006-01-02 15:04:05")))
	}
	if len(conditions) == 0 {
		return ""
	}
	return " AND " + strings.Join(conditions, " AND ")
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

	if err = m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			err = nil
		} else {
			err = fmt.Errorf("database migration failed: %w", err)
		}
	}
	return err
}
