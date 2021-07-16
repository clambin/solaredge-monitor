package store

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	log "github.com/sirupsen/logrus"
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
	psqlInfo    string
	initialized bool
}

func NewPostgresDB(host string, port int, database string, user string, password string) (handle *PostgresDB) {
	handle = &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		initialized: false,
	}

	dbh, err := sql.Open("postgres", handle.psqlInfo)
	if err != nil {
		log.WithError(err).Fatalf("failed to open database '%s'", database)
	}

	prometheus.DefaultRegisterer.MustRegister(collectors.NewDBStatsCollector(dbh, database))

	return
}

func (db *PostgresDB) initializeDB(dbh *sql.DB) (err error) {
	if db.initialized {
		return nil
	}

	_, err = dbh.Exec(`
		CREATE TABLE IF NOT EXISTS solar (
			timestamp TIMESTAMP WITHOUT TIME ZONE,
			intensity NUMERIC,
			power NUMERIC
		)
	`)

	if err == nil {
		db.initialized = true
	}

	return
}

func (db *PostgresDB) Store(measurement Measurement) (err error) {
	var dbh *sql.DB
	if dbh, err = sql.Open("postgres", db.psqlInfo); err != nil {
		return fmt.Errorf("failed to open database: %s", err.Error())
	}

	defer func(dbh *sql.DB) {
		_ = dbh.Close()
	}(dbh)

	if err = db.initializeDB(dbh); err != nil {
		return fmt.Errorf("failed to initialize database: %s", err.Error())
	}

	// Prepare the SQL statement
	stmt, _ := dbh.Prepare(`INSERT INTO solar(timestamp, intensity, power) VALUES ($1, $2, $3)`)
	_, err = stmt.Exec(measurement.Timestamp, measurement.Intensity, measurement.Power)

	if err != nil {
		err = fmt.Errorf("failed to insert measurements in database: %s", err)
	}

	return err
}

func (db *PostgresDB) Get(from, to time.Time) (measurements []Measurement, err error) {
	var dbh *sql.DB
	if dbh, err = sql.Open("postgres", db.psqlInfo); err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err.Error())
	}

	defer func(dbh *sql.DB) {
		_ = dbh.Close()
	}(dbh)

	if err = db.initializeDB(dbh); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %s", err.Error())
	}

	var rows *sql.Rows
	rows, err = dbh.Query(fmt.Sprintf(
		"SELECT timestamp, intensity, power FROM solar WHERE timestamp >= '%s' AND timestamp <= '%s' ORDER BY 1",
		from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

	if err == nil {
		defer func() {
			_ = rows.Close()
		}()
	}

	for err == nil && rows.Next() {
		var measurement Measurement
		if err = rows.Scan(&measurement.Timestamp, &measurement.Intensity, &measurement.Power); err == nil {
			measurements = append(measurements, measurement)
		}
	}

	return
}

func (db *PostgresDB) GetAll() (measurements []Measurement, err error) {
	var dbh *sql.DB
	if dbh, err = sql.Open("postgres", db.psqlInfo); err != nil {
		return nil, fmt.Errorf("failed to open database: %s", err.Error())
	}

	defer func(dbh *sql.DB) {
		_ = dbh.Close()
	}(dbh)

	if err = db.initializeDB(dbh); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %s", err.Error())
	}

	var rows *sql.Rows
	rows, err = dbh.Query("SELECT timestamp, intensity, power FROM solar ORDER BY 1")

	if err == nil {
		defer func() {
			_ = rows.Close()
		}()
	}

	for err == nil && rows.Next() {
		var measurement Measurement
		if err = rows.Scan(&measurement.Timestamp, &measurement.Intensity, &measurement.Power); err == nil {
			measurements = append(measurements, measurement)
		}
	}

	return
}
