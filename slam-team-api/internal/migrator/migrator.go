// Package migrator runs the numbered SQL migrations embedded in
// slam-team-api/migrations. There is no AutoMigrate anywhere in this project:
// the schema is those files, generated to match slamteam_db.dbml.
package migrator

import (
	"database/sql"
	"errors"
	"fmt"

	"slam-team-api/migrations"

	"github.com/golang-migrate/migrate/v4"
	pgxdriver "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// New builds a migrate instance on an existing *sql.DB, so the CLI reuses the
// pool opened by internal/database instead of dialling a second connection.
// Applied versions are tracked in the schema_migrations table.
func New(db *sql.DB) (*migrate.Migrate, error) {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("baca migrasi ter-embed: %w", err)
	}
	drv, err := pgxdriver.WithInstance(db, &pgxdriver.Config{})
	if err != nil {
		return nil, fmt.Errorf("siapkan driver migrasi: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", drv)
	if err != nil {
		return nil, fmt.Errorf("inisialisasi migrator: %w", err)
	}
	return m, nil
}

// Up applies every pending migration. It is a no-op when the schema is current.
func Up(m *migrate.Migrate) error {
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Down rolls back n migrations, or every applied migration when n <= 0.
func Down(m *migrate.Migrate, n int) error {
	var err error
	if n <= 0 {
		err = m.Down()
	} else {
		err = m.Steps(-n)
	}
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

// Version reports the current schema version. applied is false when no
// migration has run yet; dirty means a previous run failed halfway and the
// version must be fixed manually before migrating again.
func Version(m *migrate.Migrate) (version uint, dirty bool, applied bool, err error) {
	version, dirty, err = m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, false, nil
	}
	if err != nil {
		return 0, false, false, err
	}
	return version, dirty, true, nil
}
