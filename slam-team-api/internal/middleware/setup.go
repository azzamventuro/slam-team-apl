package middleware

import (
	"time"

	"slam-team-api/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global attaches process-wide middleware (panic recovery, CORS) to the engine.
// Per-route middleware (e.g. JWTAuth) is wired in each module's SetupRoutes.
//
// CORS is driven by cfg.CORS rather than cors.Default(): the browser sends a
// preflight OPTIONS for every multipart upload, so the allowed methods/headers
// must cover it. Content-Type is allowed without restricting its value —
// multipart/form-data carries a browser-generated boundary, so the request
// header is never something we can enumerate up front.
func Global(engine *gin.Engine, cfg *config.Config) {
	engine.Use(gin.Recovery())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.Origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Disposition", "Content-Length"},
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           12 * time.Hour,
	}))
}
