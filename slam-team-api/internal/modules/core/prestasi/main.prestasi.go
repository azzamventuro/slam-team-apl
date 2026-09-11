// Package prestasi wires the prestasi feature module: repository → service → handler.
//
// The module owns 5 CRUD endpoints + 1 dropdown endpoint under /prestasi/*
// and 1 public endpoint under /public/prestasi (no auth required).
package prestasi

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/prestasi/handler"
	"slam-team-api/internal/modules/core/prestasi/repository"
	"slam-team-api/internal/modules/core/prestasi/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the prestasi feature's handler and route dependencies.
type Module struct {
	h      *handler.PrestasiHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewPrestasiRepository(db)
	svc := service.NewPrestasiService(repo, auditor)
	return &Module{
		h:      handler.NewPrestasiHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the prestasi endpoints under /prestasi.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/prestasi", middleware.JWTAuth(m.jwtMgr))
	g.GET("/anggota-tersedia", m.perm.Require("prestasi.read"), m.h.AnggotaTersedia)
	g.GET("",             m.perm.Require("prestasi.read"),   m.h.List)
	g.GET("/:id",         m.perm.Require("prestasi.read"),   m.h.Detail)
	g.POST("",            m.perm.Require("prestasi.create"), m.h.Create)
	g.PUT("/:id",         m.perm.Require("prestasi.update"), m.h.Update)
	g.DELETE("/:id",      m.perm.Require("prestasi.delete"), m.h.Delete)
}

// PublicSetupRoutes mounts the public endpoint under /public/prestasi (no auth).
func (m *Module) PublicSetupRoutes(rg *gin.RouterGroup) {
	rg.GET("/prestasi", m.h.PublicList)
}
