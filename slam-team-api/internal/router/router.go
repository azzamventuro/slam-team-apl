// Package router builds the Gin engine and mounts the /api/v1 route group.
package router

import (
	"net/http"
	"time"

	"slam-team-api/internal/config"
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth"
	"slam-team-api/internal/modules/core/file"
	"slam-team-api/internal/modules/core/pengaturan"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup builds the engine, attaches global middleware, exposes /health, and
// registers every feature module under /api/v1.
//
// It returns an error when a module cannot be constructed — the file layer
// refuses to start on an unwritable STORAGE_ROOT — so a misconfiguration stops
// the process at boot instead of surfacing as a failed upload later.
func Setup(cfg *config.Config, db *gorm.DB, jwtMgr *jwt.Manager, rdb *goredis.Client) (*gin.Engine, error) {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	middleware.Global(engine, cfg)

	// Liveness/readiness probe for load balancers and uptime monitors, not a
	// business endpoint — deliberately flat JSON, outside /api/v1 and outside
	// the {success,message,data} envelope.
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"time":    time.Now().UTC().Format(time.RFC3339),
			"version": cfg.App.Version,
			"env":     cfg.App.Env,
			"redis":   rdb != nil,
		})
	})

	apiV1 := engine.Group("/api/v1")

	// Shared audit writer: every module that changes state records a
	// log_aktivitas row through it.
	auditor := audit.NewWriter(db)

	// --- Register feature modules here ---
	auth.Initialize(db, jwtMgr).SetupRoutes(apiV1)
	pengaturan.Initialize(db, jwtMgr, auditor).SetupRoutes(apiV1)

	// The file layer reads its variant sizes from mst_pengaturan, so it is
	// registered after it.
	fileModule, err := file.Initialize(db, jwtMgr, cfg, auditor)
	if err != nil {
		return nil, err
	}
	fileModule.SetupRoutes(apiV1)

	return engine, nil
}
