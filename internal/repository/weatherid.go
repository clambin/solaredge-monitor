package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

func (db *PostgresDB) GetWeatherID(weather string) (int, error) {
	var weatherID int
	row := db.DBX.QueryRow(fmt.Sprintf("SELECT id FROM weatherids WHERE weather = '%s'", weather))
	err := row.Scan(&weatherID)
	if errors.Is(err, sql.ErrNoRows) {
		slog.Debug("defining new weather type", "weather", weather)
		if _, err = db.DBX.Exec(fmt.Sprintf("INSERT INTO weatherids(id, weather) VALUES(nextval('weatherid'), '%s')", weather)); err == nil {
			weatherID, err = db.GetWeatherID(weather)
		}
	}
	return weatherID, err
}

func (db *PostgresDB) GetWeather(id int) (string, error) {
	var weather string
	row := db.DBX.QueryRow("SELECT weather FROM weatherids WHERE id = $1", id)
	err := row.Scan(&weather)
	return weather, err
}
