// Package jadwal wires the jadwal feature module: repository → service → handler.
//
// The module owns the schedule definition (jadwal) and its materialised
// sessions (jadwal_sesi): 5 CRUD endpoints under /jadwal/*, the session
// generator and session list under /jadwal/:id/*, and session cancellation
// under /sesi/:id/batalkan.
package jadwal

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/jadwal/handler"
	"slam-team-api/internal/modules/core/jadwal/repository"
	"slam-team-api/internal/modules/core/jadwal/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the jadwal feature's handler and route dependencies.
type Module struct {
	h      *handler.JadwalHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewJadwalRepository(db)
	svc := service.NewJadwalService(repo, auditor)
	return &Module{
		h:      handler.NewJadwalHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the jadwal endpoints. Permission bands (seeded in
// slamctl seed rbac): jadwal.read = SA/Admin/Mod/User; create/update =
// SA/Admin/Mod; delete = SA/Admin; batal_sesi = SA/Admin/Mod.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/jadwal", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("jadwal.read"), m.h.List)
	g.GET("/:id", m.perm.Require("jadwal.read"), m.h.Detail)
	g.POST("", m.perm.Require("jadwal.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("jadwal.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("jadwal.delete"), m.h.Delete)

	// Materialising sessions changes the schedule's state → jadwal.update.
	g.POST("/:id/generate-sesi", m.perm.Require("jadwal.update"), m.h.GenerateSesi)
	g.GET("/:id/sesi", m.perm.Require("jadwal.read"), m.h.ListSesi)

	// Cancelling a session affects every participant → its own permission.
	s := rg.Group("/sesi", middleware.JWTAuth(m.jwtMgr))
	s.PATCH("/:id/batalkan", m.perm.Require("jadwal.batal_sesi"), m.h.BatalkanSesi)
}
