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
	psqlInfo string
	database string
	dbh      *sql.DB
}

func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	return &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		database: database,
	}
}

func (db *PostgresDB) Store(measurement Measurement) (err error) {
	db.initialize()

	// Prepare the SQL statement
	stmt, _ := db.dbh.Prepare(`INSERT INTO solar(timestamp, intensity, power) VALUES ($1, $2, $3)`)
	_, err = stmt.Exec(measurement.Timestamp, measurement.Intensity, measurement.Power)

	if err != nil {
		err = fmt.Errorf("failed to insert measurements in database: %s", err)
		db.close()
	}

	return err
}

func (db *PostgresDB) Get(from, to time.Time) (measurements []Measurement, err error) {
	db.initialize()

	var rows *sql.Rows
	rows, err = db.dbh.Query(fmt.Sprintf(
		"SELECT timestamp, intensity, power FROM solar WHERE timestamp >= '%s' AND timestamp <= '%s' ORDER BY 1",
		from.Format("2006-01-02 15:04:05"), to.Format("2006-01-02 15:04:05")))

	if err != nil {
		db.close()
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
	db.initialize()

	var rows *sql.Rows
	rows, err = db.dbh.Query("SELECT timestamp, intensity, power FROM solar ORDER BY 1")

	if err != nil {
		db.close()
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

func (db *PostgresDB) initialize() {
	if db.dbh != nil {
		return
	}

	var err error
	db.dbh, err = sql.Open("postgres", db.psqlInfo)
	if err != nil {
		log.WithError(err).Fatalf("failed to open database '%s'", db.database)
	}

	prometheus.DefaultRegisterer.MustRegister(collectors.NewDBStatsCollector(db.dbh, db.database))

	_, err = db.dbh.Exec(`
		CREATE TABLE IF NOT EXISTS solar (
			timestamp TIMESTAMP WITHOUT TIME ZONE,
			intensity NUMERIC,
			power NUMERIC
		)
	`)

	if err != nil {
		log.WithError(err).Fatalf("unable to intialize database '%s'", db.database)
	}

	return
}

func (db *PostgresDB) close() {
	if db.dbh != nil {
		_ = db.dbh.Close()
	}
	db.dbh = nil
}
