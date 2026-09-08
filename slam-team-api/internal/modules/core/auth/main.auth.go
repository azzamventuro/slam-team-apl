// Package auth is the authentication feature module. It follows the standard
// module layout: domain / dto / repository / service / handler, wired here.
package auth

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/handler"
	"slam-team-api/internal/modules/core/auth/repository"
	"slam-team-api/internal/modules/core/auth/service"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the auth feature's handler and its route dependencies.
type Module struct {
	handler *handler.AuthHandler
	jwtMgr  *jwt.Manager
}

// Initialize composes repository → service → handler.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager) *Module {
	repo := repository.NewUserRepository(db)
	svc := service.NewAuthService(repo, jwtMgr)
	return &Module{handler: handler.NewAuthHandler(svc), jwtMgr: jwtMgr}
}

// SetupRoutes mounts the module's endpoints under the given group.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/auth")
	grp.POST("/login", m.handler.Login)
	grp.GET("/me", middleware.JWTAuth(m.jwtMgr), m.handler.Me)
}
