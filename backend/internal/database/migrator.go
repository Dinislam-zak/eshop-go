package database

import (
	"embed"
	"errors"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	subFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	sourceInstance, err := iofs.New(subFS, ".")
	if err != nil {
		return err
	}

	dbInstance, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", dbInstance)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}