// Package auth is the authentication feature module. It follows the standard
// module layout: domain / dto / repository / service / handler, wired here.
package auth

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/handler"
	"slam-team-api/internal/modules/core/auth/repository"
	"slam-team-api/internal/modules/core/auth/service"
	hakaksesrepo "slam-team-api/internal/modules/core/hakakses/repository"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the auth feature's handler and its route dependencies.
type Module struct {
	handler  *handler.AuthHandler
	jwtMgr   *jwt.Manager
	permGuard *middleware.PermGuard
}

// Initialize composes repository → service → handler.
// permGuard may be nil if the hakakses module hasn't been wired yet.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, hakaksesRepo *hakaksesrepo.HakaksesRepository, permGuard *middleware.PermGuard) *Module {
	repo := repository.NewUserRepository(db)
	svc := service.NewAuthService(repo, jwtMgr, hakaksesRepo, permGuard)
	return &Module{handler: handler.NewAuthHandler(svc), jwtMgr: jwtMgr, permGuard: permGuard}
}

// SetupRoutes mounts the module's endpoints under the given group.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/auth")
	grp.POST("/login", m.handler.Login)
	grp.GET("/me", middleware.JWTAuth(m.jwtMgr), m.handler.Me)
	grp.GET("/me/permissions", middleware.JWTAuth(m.jwtMgr), m.handler.MePermissions)
}
