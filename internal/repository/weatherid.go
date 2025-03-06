package repository

import (
	"database/sql"
	"errors"
	"log/slog"
)

func (db *PostgresDB) GetWeatherID(weather string) (int, error) {
	var weatherID int
	err := db.DBX.Get(&weatherID, "SELECT id FROM weatherids WHERE weather = $1", weather)
	if !errors.Is(err, sql.ErrNoRows) {
		return weatherID, err
	}
	slog.Debug("defining new weather type", "weather", weather)
	if _, err = db.DBX.Exec("INSERT INTO weatherids(id, weather) VALUES(nextval('weatherid'), $1)", weather); err == nil {
		weatherID, err = db.GetWeatherID(weather)
	}
	return weatherID, err
}

func (db *PostgresDB) GetWeather(id int) (string, error) {
	var weather string
	row := db.DBX.QueryRow("SELECT weather FROM weatherids WHERE id = $1", id)
	err := row.Scan(&weather)
	return weather, err
}
