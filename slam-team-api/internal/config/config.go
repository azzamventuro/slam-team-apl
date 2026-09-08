package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is the fully-typed application configuration, loaded from the
// environment (and an optional .env file) at startup.
type Config struct {
	App   AppConfig
	DB    DBConfig
	JWT   JWTConfig
	Redis RedisConfig
}

type AppConfig struct {
	Env  string // development | production
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	SSLMode  string // disable | require | verify-full ...
}

type JWTConfig struct {
	Secret string
	Issuer string
	TTL    time.Duration
}

type RedisConfig struct {
	Addr     string // empty → Redis disabled
	Password string
	DB       int
}

// Load reads configuration from the environment. A .env file is loaded when
// present but is optional.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		App: AppConfig{
			Env:  getenv("APP_ENV", "development"),
			Port: getenv("APP_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getenv("DB_HOST", "127.0.0.1"),
			Port:     getenv("DB_PORT", "5432"),
			Name:     getenv("DB_NAME", "slamteam_db"),
			User:     getenv("DB_USER", "postgres"),
			Password: getenv("DB_PASSWORD", ""),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret: getenv("JWT_SECRET", "change-me"),
			Issuer: getenv("JWT_ISSUER", "slam-team-api"),
			TTL:    time.Duration(getenvInt("JWT_TTL_HOURS", 24)) * time.Hour,
		},
		Redis: RedisConfig{
			Addr:     getenv("REDIS_ADDR", ""),
			Password: getenv("REDIS_PASSWORD", ""),
			DB:       getenvInt("REDIS_DB", 0),
		},
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
