// Package anggota wires the anggota feature module: repository → service → handler.
//
// The module owns 5 endpoints under /anggota/* for CRUD operations on
// Master Data Anggota (core member records).
package anggota

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/anggota/handler"
	"slam-team-api/internal/modules/core/anggota/repository"
	"slam-team-api/internal/modules/core/anggota/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the anggota feature's handler and route dependencies.
type Module struct {
	h      *handler.AnggotaHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewAnggotaRepository(db)
	svc := service.NewAnggotaService(repo, auditor)
	return &Module{
		h:      handler.NewAnggotaHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the 5 anggota endpoints under /anggota.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/anggota", middleware.JWTAuth(m.jwtMgr))
	g.GET("",        m.perm.Require("anggota.read"),   m.h.List)
	g.GET("/:id",    m.perm.Require("anggota.read"),   m.h.Detail)
	g.POST("",       m.perm.Require("anggota.create"), m.h.Create)
	g.PUT("/:id",    m.perm.Require("anggota.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("anggota.delete"), m.h.Delete)
}
