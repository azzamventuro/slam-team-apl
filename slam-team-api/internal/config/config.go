package config

import (
	"os"
	"strconv"
	"strings"
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
	CORS  CORSConfig
	File  FileConfig
}

type AppConfig struct {
	Env      string // development | production
	Port     string
	Version  string // build version; override with -ldflags
	Timezone string // default IANA zone for users/schedules
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
	Secret     string
	Issuer     string
	TTL        time.Duration
	RefreshTTL time.Duration
}

type RedisConfig struct {
	Addr     string // empty → Redis disabled
	Password string
	DB       int
}

// CORSConfig drives the browser CORS policy. Origins is an explicit allowlist
// (never "*" when AllowCredentials is on).
type CORSConfig struct {
	Origins          []string
	AllowCredentials bool
}

// FileConfig holds the upload layer's limits and where the bytes live.
type FileConfig struct {
	MaxUploadMB int
	// StorageRoot is the directory the disk layout hangs under:
	// {StorageRoot}/{kategori}/{tahun}/{bulan}/{uuid}/{varian}.{ext}. It must
	// be writable by the API process — an unwritable root is the classic
	// "upload reported success but the file is gone" (rancangan Bab 6.6).
	StorageRoot string
}

// Load reads configuration from the environment. A .env file is loaded when
// present but is optional.
func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		App: AppConfig{
			Env:      getenv("APP_ENV", "development"),
			Port:     getenv("APP_PORT", "8080"),
			Version:  getenv("APP_VERSION", "dev"),
			Timezone: getenv("APP_TIMEZONE", "Asia/Jakarta"),
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
			Secret:     getenv("JWT_SECRET", "change-me"),
			Issuer:     getenv("JWT_ISSUER", "slam-team-api"),
			TTL:        time.Duration(getenvInt("JWT_TTL_HOURS", 24)) * time.Hour,
			RefreshTTL: time.Duration(getenvInt("JWT_REFRESH_TTL_HOURS", 720)) * time.Hour,
		},
		Redis: RedisConfig{
			Addr:     getenv("REDIS_ADDR", ""),
			Password: getenv("REDIS_PASSWORD", ""),
			DB:       getenvInt("REDIS_DB", 0),
		},
		CORS: CORSConfig{
			Origins:          getenvList("CORS_ORIGINS", []string{"http://localhost:4200"}),
			AllowCredentials: getenvBool("CORS_ALLOW_CREDENTIALS", true),
		},
		File: FileConfig{
			MaxUploadMB: getenvInt("FILE_MAX_UPLOAD_MB", 15),
			StorageRoot: getenv("STORAGE_ROOT", "./storage"),
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

func getenvBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

// getenvList reads a comma-separated list, trimming blanks around each entry.
func getenvList(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return def
	}
	return out
}
