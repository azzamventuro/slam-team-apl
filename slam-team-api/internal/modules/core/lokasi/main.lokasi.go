// Package lokasi wires the lokasi feature module: repository → service → handler.
//
// The module owns 5 endpoints under /lokasi/* for CRUD operations on
// Master Data Lokasi (training/activity locations with geofence + timezone).
package lokasi

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/lokasi/handler"
	"slam-team-api/internal/modules/core/lokasi/repository"
	"slam-team-api/internal/modules/core/lokasi/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the lokasi feature's handler and route dependencies.
type Module struct {
	h      *handler.LokasiHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewLokasiRepository(db)
	svc := service.NewLokasiService(repo, auditor)
	return &Module{
		h:      handler.NewLokasiHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the 5 lokasi endpoints under /lokasi.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/lokasi", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("lokasi.read"), m.h.List)
	g.GET("/:id", m.perm.Require("lokasi.read"), m.h.Detail)
	g.POST("", m.perm.Require("lokasi.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("lokasi.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("lokasi.delete"), m.h.Delete)
}
