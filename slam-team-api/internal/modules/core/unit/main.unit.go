// Package unit wires the unit feature module: repository → service → handler.
//
// The module owns 5 CRUD endpoints + 1 dropdown endpoint under /unit/* for
// Master Data Unit (airsoft gun registry).
package unit

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/unit/handler"
	"slam-team-api/internal/modules/core/unit/repository"
	"slam-team-api/internal/modules/core/unit/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the unit feature's handler and route dependencies.
type Module struct {
	h      *handler.UnitHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewUnitRepository(db)
	svc := service.NewUnitService(repo, auditor)
	return &Module{
		h:      handler.NewUnitHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the unit endpoints under /unit.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/unit", middleware.JWTAuth(m.jwtMgr))
	g.GET("",             m.perm.Require("unit.read"),   m.h.List)
	g.GET("/anggota-tersedia", m.perm.Require("unit.read"), m.h.AnggotaTersedia)
	g.GET("/:id",         m.perm.Require("unit.read"),   m.h.Detail)
	g.POST("",            m.perm.Require("unit.create"), m.h.Create)
	g.PUT("/:id",         m.perm.Require("unit.update"), m.h.Update)
	g.DELETE("/:id",      m.perm.Require("unit.delete"), m.h.Delete)
}
