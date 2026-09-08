package middleware

import (
	"slam-team-api/internal/config"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Global attaches process-wide middleware (panic recovery, CORS) to the engine.
// Per-route middleware (e.g. JWTAuth) is wired in each module's SetupRoutes.
func Global(engine *gin.Engine, cfg *config.Config) {
	engine.Use(gin.Recovery())
	engine.Use(cors.Default())
}
