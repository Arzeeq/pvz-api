package pg

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type Migrator struct {
	migrationsDir string
	connectionStr string
}

func NewMigrator(migrationsDir, connectionStr string) *Migrator {
	return &Migrator{migrationsDir: migrationsDir, connectionStr: connectionStr}
}

func (m *Migrator) Up() error {
	mig, err := m.initMigrate()
	if err != nil {
		return err
	}
	defer mig.Close()

	if err = mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

func (m *Migrator) Down() error {
	mig, err := m.initMigrate()
	if err != nil {
		return err
	}
	defer mig.Close()

	if err := mig.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	return nil
}

func (m *Migrator) initMigrate() (*migrate.Migrate, error) {
	if _, err := os.Stat(m.migrationsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("migrations directory does not exist: %w", err)
	}

	source, err := iofs.New(os.DirFS(m.migrationsDir), ".")
	if err != nil {
		return nil, fmt.Errorf("failed to create migrations source: %w", err)
	}

	mig, err := migrate.NewWithSourceInstance("iofs", source, m.connectionStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	return mig, nil
}
