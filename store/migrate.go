package store

import (
	"embed"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*
var migrations embed.FS

func (db *PostgresDB) migrate() (err error) {
	var src source.Driver
	if src, err = iofs.New(migrations, "migrations"); err != nil {
		return fmt.Errorf("invalid migration source: %w", err)
	}

	var dbDriver database.Driver
	if dbDriver, err = postgres.WithInstance(db.DBH, &postgres.Config{DatabaseName: "solaredge"}); err != nil {
		return fmt.Errorf("invalid migration target: %w", err)
	}

	var m *migrate.Migrate
	if m, err = migrate.NewWithInstance("migrations", src, "solaredge", dbDriver); err != nil {
		return fmt.Errorf("unable to migrate database: %w", err)
	}

	if err = m.Up(); err == migrate.ErrNoChange {
		err = nil
	}

	return
}
