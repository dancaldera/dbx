package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/dancaldera/mirador/internal/models"
)

// ValidateConnectionString validates the format of connection strings
func ValidateConnectionString(driver, connectionStr string) error {
	if connectionStr == "" {
		return fmt.Errorf("connection string cannot be empty")
	}

	switch driver {
	case "postgres":
		if !strings.HasPrefix(connectionStr, "postgres://") && !strings.HasPrefix(connectionStr, "postgresql://") {
			return fmt.Errorf("PostgreSQL connection string should start with 'postgres://' or 'postgresql://'")
		}
	case "mysql":
		if strings.HasPrefix(connectionStr, "mysql://") {
			return fmt.Errorf("MySQL connection string should not include 'mysql://' prefix. Use format: user:password@tcp(host:port)/dbname")
		}
	case "sqlite3":
		return ValidateSQLiteConnection(connectionStr)
	default:
		return fmt.Errorf("unsupported database driver: %s", driver)
	}

	return nil
}

// ValidateSQLiteConnection validates SQLite database file
func ValidateSQLiteConnection(path string) error {
	if path == "" {
		return fmt.Errorf("SQLite database path cannot be empty")
	}

	if strings.Contains(path, "..") {
		return fmt.Errorf("relative paths with '..' are not allowed for security reasons")
	}

	if !strings.HasSuffix(path, ".db") && !strings.HasSuffix(path, ".sqlite") && !strings.HasSuffix(path, ".sqlite3") {
		return fmt.Errorf("SQLite file should have .db, .sqlite, or .sqlite3 extension")
	}

	return nil
}

// TestConnectionWithTimeout tests a database connection with timeout
func TestConnectionWithTimeout(driver, connectionStr string) models.TestConnectionResult {
	timeout := 10 * time.Second
	done := make(chan models.TestConnectionResult, 1)

	go func() {
		db, err := sql.Open(driver, connectionStr)
		if err != nil {
			done <- models.TestConnectionResult{Success: false, Err: err}
			return
		}
		defer db.Close()

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		err = db.PingContext(ctx)
		if err != nil {
			done <- models.TestConnectionResult{Success: false, Err: err}
			return
		}

		done <- models.TestConnectionResult{Success: true, Err: nil}
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(timeout):
		return models.TestConnectionResult{
			Success: false,
			Err:     fmt.Errorf("connection timeout after %v", timeout),
		}
	}
}
