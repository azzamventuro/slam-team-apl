// Package instansi wires the instansi feature module: repository → service → handler.
//
// The module owns 5 endpoints under /instansi/* for CRUD operations on
// Master Data Instansi / Sekolah.
package instansi

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/instansi/handler"
	"slam-team-api/internal/modules/core/instansi/repository"
	"slam-team-api/internal/modules/core/instansi/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the instansi feature's handler and route dependencies.
type Module struct {
	h      *handler.InstansiHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewInstansiRepository(db)
	svc := service.NewInstansiService(repo, auditor)
	return &Module{
		h:      handler.NewInstansiHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the 5 instansi endpoints under /instansi.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/instansi", middleware.JWTAuth(m.jwtMgr))
	g.GET("",        m.perm.Require("instansi.read"),   m.h.List)
	g.GET("/:id",    m.perm.Require("instansi.read"),   m.h.Detail)
	g.POST("",       m.perm.Require("instansi.create"), m.h.Create)
	g.PUT("/:id",    m.perm.Require("instansi.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("instansi.delete"), m.h.Delete)
}
