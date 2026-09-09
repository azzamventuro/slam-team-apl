package main

import (
	"database/sql"
	"fmt"

	"slam-team-api/internal/config"
	"slam-team-api/internal/database"

	"gorm.io/gorm"
)

// openDB loads config and opens the same pooled connection cmd/api uses.
func openDB() (*config.Config, *gorm.DB, error) {
	cfg := config.Load()
	db, err := database.New(cfg.DB)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"tidak bisa terhubung ke PostgreSQL %s:%s/%s — periksa DB_* di .env (%w)",
			cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, err,
		)
	}
	return cfg, db, nil
}

// closeDB releases the pool; ignore errors, the process is exiting.
func closeDB(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func sqlDB(db *gorm.DB) (*sql.DB, error) {
	s, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ambil koneksi sql: %w", err)
	}
	return s, nil
}

// tableExists reports whether a table is present in the current schema. Used to
// tell "migrations have not run yet" apart from a genuine query failure.
func tableExists(db *gorm.DB, table string) (bool, error) {
	var exists bool
	err := db.Raw("SELECT to_regclass(?) IS NOT NULL", "public."+table).Scan(&exists).Error
	return exists, err
}
