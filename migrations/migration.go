package migrations

import (
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

func RunMigrations(postgresURL, migrationDir string) error {
	db, err := goose.OpenDBWithDriver("postgres", postgresURL)
	if err != nil {
		return fmt.Errorf("failed to open DB: %v", err)
	}
	defer db.Close()

	if err := goose.Up(db, migrationDir); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}
	return nil
}
