package store

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"time"
)

type DB interface {
	Store(Measurement) error
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

func NewPostgresDB(host string, port int, database string, user string, password string) *PostgresDB {
	return &PostgresDB{
		psqlInfo: fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, database),
		initialized: false,
	}
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

	dbh, err = sql.Open("postgres", db.psqlInfo)

	if err != nil {
		return fmt.Errorf("failed to open database: %s", err.Error())
	}

	defer func(dbh *sql.DB) {
		_ = dbh.Close()
	}(dbh)

	err = db.initializeDB(dbh)

	if err != nil {
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
