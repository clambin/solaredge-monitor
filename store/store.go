package store

import (
	"database/sql"
	"fmt"
	// postgres driver for database/sql
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
	DBH      *sql.DB
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
	// Prepare the SQL statement
	stmt, _ := db.DBH.Prepare(`INSERT INTO solar(timestamp, intensity, power) VALUES ($1, $2, $3)`)
	_, err = stmt.Exec(measurement.Timestamp, measurement.Intensity, measurement.Power)

	return err
}

func (db *PostgresDB) Get(from, to time.Time) (measurements []Measurement, err error) {
	var rows *sql.Rows
	rows, err = db.DBH.Query(fmt.Sprintf(
		"SELECT timestamp, intensity, power FROM solar WHERE timestamp >= '%s' AND timestamp <= '%s' ORDER BY 1",
		from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

	if err != nil {
		return nil, err
	}

	for err == nil && rows.Next() {
		var measurement Measurement
		if err = rows.Scan(&measurement.Timestamp, &measurement.Intensity, &measurement.Power); err == nil {
			measurements = append(measurements, measurement)
		}
	}
	_ = rows.Close()

	return
}

func (db *PostgresDB) GetAll() (measurements []Measurement, err error) {
	var rows *sql.Rows
	rows, err = db.DBH.Query("SELECT timestamp, intensity, power FROM solar ORDER BY 1")

	if err != nil {
		return nil, err
	}

	for err == nil && rows.Next() {
		var measurement Measurement
		if err = rows.Scan(&measurement.Timestamp, &measurement.Intensity, &measurement.Power); err == nil {
			measurements = append(measurements, measurement)
		}
	}
	_ = rows.Close()

	return
}

func (db *PostgresDB) initialize() (err error) {
	if db.DBH, err = sql.Open("postgres", db.psqlInfo); err == nil {
		err = db.DBH.Ping()
	}
	if err != nil {
		return fmt.Errorf("database open: %w", err)
	}

	prometheus.DefaultRegisterer.MustRegister(collectors.NewDBStatsCollector(db.DBH, db.database))

	if err = db.migrate(); err != nil {
		err = fmt.Errorf("database initialization: %w", err)
	}
	return err
}
