package store

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/exp/slog"
)

func (db *PostgresDB) GetWeatherID(weather string) (int, error) {
	var weatherID int
	row := db.DBH.QueryRow(fmt.Sprintf("SELECT id FROM weatherIDs WHERE weather = '%s'", weather))
	err := row.Scan(&weatherID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Debug("defining new weather type", "weather", weather)
		if _, err = db.DBH.Exec(fmt.Sprintf("INSERT INTO weatherIDs(id, weather) VALUES(nextval('weatherid'), '%s')", weather)); err == nil {
			weatherID, err = db.GetWeatherID(weather)
		}
	}
	return weatherID, err
}

func (db *PostgresDB) GetWeather(id int) (string, error) {
	var weather string
	row := db.DBH.QueryRow("SELECT weather FROM weatherids WHERE id = $1", id)
	err := row.Scan(&weather)
	return weather, err
}
