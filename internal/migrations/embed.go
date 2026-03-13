package migrations

import (
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed postgres/*.sql
var migrationsFiles embed.FS

func Exec(dbUrl string) error {
	source, err := iofs.New(migrationsFiles, "postgres")

	if err != nil {
		return fmt.Errorf("failed to create migrations source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, dbUrl)
	if err != nil {
		return fmt.Errorf("failed to create migrations: %w", err)
	}
	defer m.Close()

	if err = m.Up(); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
