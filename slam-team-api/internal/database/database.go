// Package database opens the application's PostgreSQL connection via GORM.
package database

import (
	"fmt"
	"time"

	"slam-team-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// New opens a pooled GORM/PostgreSQL connection and verifies it with a ping.
func New(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}
