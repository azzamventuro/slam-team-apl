// Package profile_club wires the profile club feature module: repository → service → handler.
//
// The module owns CRUD endpoints under /profile-club/* (admin, bearer auth)
// and one public endpoint under /public/profil (no auth required).
package profileclub

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/profileclub/handler"
	"slam-team-api/internal/modules/core/profileclub/repository"
	"slam-team-api/internal/modules/core/profileclub/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the profile club feature's handler and route dependencies.
type Module struct {
	h      *handler.ProfileClubHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
// The repository implements FileExistChecker via its FileExists method.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, aw *audit.Writer) *Module {
	repo := repository.NewProfileClubRepository(db)
	svc := service.NewProfileClubService(repo, repo, aw)
	return &Module{
		h:      handler.NewProfileClubHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the profile_club endpoints under /profile-club.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/profile-club", middleware.JWTAuth(m.jwtMgr))
	g.GET("",      m.perm.Require("profile_club.read"),   m.h.List)
	g.GET("/:id",  m.perm.Require("profile_club.read"),   m.h.Detail)
	g.POST("",     m.perm.Require("profile_club.create"), m.h.Create)
	g.PUT("/:id",  m.perm.Require("profile_club.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("profile_club.delete"), m.h.Delete)
}

// PublicSetupRoutes mounts the public endpoint under /public/profil (no auth).
func (m *Module) PublicSetupRoutes(rg *gin.RouterGroup) {
	rg.GET("/profil", m.h.PublicProfil)
}
