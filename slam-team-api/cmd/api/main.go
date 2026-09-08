package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"slam-team-api/internal/config"
	"slam-team-api/internal/database"
	"slam-team-api/internal/router"
	"slam-team-api/internal/shared/redis"
	"slam-team-api/pkg/jwt"
	"slam-team-api/pkg/logger"
	"slam-team-api/pkg/validator"

	"go.uber.org/zap"
)

// main wires dependencies in order, then serves with graceful shutdown.
//
//	1. config.Load()      → env vars into typed structs
//	2. logger.Initialize()→ Zap (JSON in prod, console+color in dev)
//	3. validator.Register()→ custom go-playground rules on Gin
//	4. jwt.New()          → HS256 issuer/verifier
//	5. redis.Initialize() → optional, non-fatal if unavailable
//	6. database.New()     → GORM MySQL pool (fatal on failure)
//	7. router.Setup()     → mounts /api/v1 and registers modules
func main() {
	cfg := config.Load()

	logger.Initialize(cfg.App.Env)
	defer logger.Sync()

	validator.Register()

	jwtMgr := jwt.New(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.TTL)

	rdb := redis.Initialize(cfg.Redis)

	db, err := database.New(cfg.DB)
	if err != nil {
		logger.Fatal("database connect failed", zap.Error(err))
	}

	engine := router.Setup(cfg, db, jwtMgr, rdb)

	srv := &http.Server{Addr: ":" + cfg.App.Port, Handler: engine}

	go func() {
		logger.Info("server starting",
			zap.String("port", cfg.App.Port),
			zap.String("env", cfg.App.Env),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("forced shutdown", zap.Error(err))
	}
}
