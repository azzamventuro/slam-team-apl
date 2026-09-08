// Package logger is a thin Zap wrapper: JSON output in production, coloured
// console output in development. Call Initialize once at startup.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

// Initialize builds the global logger for the given environment.
func Initialize(env string) {
	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	log = l
}

// l returns the configured logger, falling back to a dev logger if Initialize
// was never called (e.g. in tests).
func l() *zap.Logger {
	if log == nil {
		log, _ = zap.NewDevelopment()
	}
	return log
}

func Info(msg string, fields ...zap.Field)  { l().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)  { l().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field) { l().Error(msg, fields...) }
func Fatal(msg string, fields ...zap.Field) { l().Fatal(msg, fields...) }

// Sync flushes buffered log entries. Call via defer in main.
func Sync() { _ = l().Sync() }
