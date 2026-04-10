package db

import (
	"database/sql"
	"errors"
	"fmt"

	appmigrations "github-release-notification-api/migrations"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func RunMigrations(database *sql.DB) error {
	sourceDriver, err := iofs.New(appmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	dbDriver, err := pgxmigrate.WithInstance(database, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("create migration database driver: %w", err)
	}

	migrationRunner, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("create migration runner: %w", err)
	}

	if err := migrationRunner.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
