// Package auth is the authentication feature module: login (username or
// email), refresh-token rotation, logout, and the identity reads (/me,
// /me/permissions). It follows the standard module layout: domain / dto /
// repository / service / handler, wired here.
package auth

import (
	"time"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/handler"
	"slam-team-api/internal/modules/core/auth/repository"
	"slam-team-api/internal/modules/core/auth/service"
	hakaksesrepo "slam-team-api/internal/modules/core/hakakses/repository"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the auth feature's handler and its route dependencies.
type Module struct {
	handler *handler.AuthHandler
	jwtMgr  *jwt.Manager
}

// Initialize composes repository → service → handler. refreshTTL is the
// lifetime of a sesi_login row (config JWT_REFRESH_TTL_HOURS).
func Initialize(
	db *gorm.DB,
	jwtMgr *jwt.Manager,
	hakaksesRepo *hakaksesrepo.HakaksesRepository,
	permGuard *middleware.PermGuard,
	auditor *audit.Writer,
	refreshTTL time.Duration,
) *Module {
	repo := repository.NewUserRepository(db)
	svc := service.NewAuthService(repo, jwtMgr, hakaksesRepo, permGuard, auditor, refreshTTL)
	return &Module{handler: handler.NewAuthHandler(svc), jwtMgr: jwtMgr}
}

// SetupRoutes mounts the module's endpoints under the given group.
// login and refresh are public; the rest require a bearer token.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/auth")
	grp.POST("/login", m.handler.Login)
	grp.POST("/refresh", m.handler.Refresh)

	bearer := grp.Group("", middleware.JWTAuth(m.jwtMgr))
	bearer.POST("/logout", m.handler.Logout)
	bearer.GET("/me", m.handler.Me)
	bearer.GET("/me/permissions", m.handler.MePermissions)
}
