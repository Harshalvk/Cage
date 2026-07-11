package store

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func runTestMigrations(databaseURL string) error {
	if err := waitForDb(databaseURL, 30*time.Second); err != nil {
		return fmt.Errorf("database never become reachable: %w", err)
	}

	m, err := migrate.New("file://../../migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to init migrate: %w", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("migrate source close error: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("migrate db close error: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

func waitForDb(databaseURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		db, err := sql.Open("pgx", databaseURL)
		if err == nil {
			pingErr := db.Ping()
			defer func() {
				if err := db.Close(); err != nil {
					log.Printf("db close error: %v", err)
				}
			}()
			if pingErr == nil {
				return nil
			}
			lastErr = pingErr
		} else {
			lastErr = err
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for database: %w", lastErr)
}
