// Package router builds the Gin engine and mounts the /api/v1 route group.
package router

import (
	"net/http"

	"slam-team-api/internal/config"
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup builds the engine, attaches global middleware, exposes /health, and
// registers every feature module under /api/v1.
func Setup(cfg *config.Config, db *gorm.DB, jwtMgr *jwt.Manager, rdb *goredis.Client) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	middleware.Global(engine, cfg)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "redis": rdb != nil})
	})

	apiV1 := engine.Group("/api/v1")

	// --- Register feature modules here ---
	auth.Initialize(db, jwtMgr).SetupRoutes(apiV1)

	return engine
}
