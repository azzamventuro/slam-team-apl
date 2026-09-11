// Package usermgmt wires the user-management feature module: repository →
// service → handler. Three surfaces (admin / moderator / user) share one
// service instance and are differentiated by the role-level band passed to
// each handler instance.
//
// Routes per surface:
//
//	GET    /{surface}                    — list (paginated, filtered by band)
//	GET    /{surface}/:id                — detail
//	POST   /{surface}                    — create
//	PUT    /{surface}/:id                — update
//	DELETE /{surface}/:id                — soft-delete
//	GET    /{surface}/peran-tersedia     — available roles for dropdown
//	GET    /{surface}/anggota-tersedia   — available anggota for dropdown
package usermgmt

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/usermgmt/handler"
	"slam-team-api/internal/modules/core/usermgmt/repository"
	"slam-team-api/internal/modules/core/usermgmt/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the user-management feature's handler and route dependencies.
type Module struct {
	adminH     *handler.UsermgmtHandler
	moderatorH *handler.UsermgmtHandler
	userH      *handler.UsermgmtHandler
	jwtMgr     *jwt.Manager
	perm       *middleware.PermGuard
}

// Initialize composes repository → service → handler × 3 surfaces and returns
// the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewUsermgmtRepository(db)
	svc := service.NewUsermgmtService(repo, auditor)

	return &Module{
		adminH:     handler.NewUsermgmtHandler(svc, service.BandAdmin),
		moderatorH: handler.NewUsermgmtHandler(svc, service.BandModerator),
		userH:      handler.NewUsermgmtHandler(svc, service.BandUser),
		jwtMgr:     jwtMgr,
		perm:       perm,
	}
}

// SetupRoutes mounts the 3 surface route groups, each with its own permission
// guard prefix. Every surface shares the same CRUD shape but is filtered by
// role-level band.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	// ── /admin surface ──
	admin := rg.Group("/admin", middleware.JWTAuth(m.jwtMgr))
	admin.GET("",             m.perm.Require("admin.read"),   m.adminH.List)
	admin.GET("/peran-tersedia",     m.perm.Require("admin.create"), m.adminH.AvailableRoles)
	admin.GET("/anggota-tersedia",   m.perm.Require("admin.create"), m.adminH.AvailableAnggota)
	admin.GET("/:id",         m.perm.Require("admin.read"),   m.adminH.Detail)
	admin.POST("",            m.perm.Require("admin.create"), m.adminH.Create)
	admin.PUT("/:id",         m.perm.Require("admin.update"), m.adminH.Update)
	admin.DELETE("/:id",      m.perm.Require("admin.delete"), m.adminH.Delete)

	// ── /moderator surface ──
	moderator := rg.Group("/moderator", middleware.JWTAuth(m.jwtMgr))
	moderator.GET("",             m.perm.Require("moderator.read"),   m.moderatorH.List)
	moderator.GET("/peran-tersedia",     m.perm.Require("moderator.create"), m.moderatorH.AvailableRoles)
	moderator.GET("/anggota-tersedia",   m.perm.Require("moderator.create"), m.moderatorH.AvailableAnggota)
	moderator.GET("/:id",         m.perm.Require("moderator.read"),   m.moderatorH.Detail)
	moderator.POST("",            m.perm.Require("moderator.create"), m.moderatorH.Create)
	moderator.PUT("/:id",         m.perm.Require("moderator.update"), m.moderatorH.Update)
	moderator.DELETE("/:id",      m.perm.Require("moderator.delete"), m.moderatorH.Delete)

	// ── /user surface ──
	user := rg.Group("/user", middleware.JWTAuth(m.jwtMgr))
	user.GET("",             m.perm.Require("user.read"),   m.userH.List)
	user.GET("/peran-tersedia",     m.perm.Require("user.create"), m.userH.AvailableRoles)
	user.GET("/anggota-tersedia",   m.perm.Require("user.create"), m.userH.AvailableAnggota)
	user.GET("/:id",         m.perm.Require("user.read"),   m.userH.Detail)
	user.POST("",            m.perm.Require("user.create"), m.userH.Create)
	user.PUT("/:id",         m.perm.Require("user.update"), m.userH.Update)
	user.DELETE("/:id",      m.perm.Require("user.delete"), m.userH.Delete)
}
