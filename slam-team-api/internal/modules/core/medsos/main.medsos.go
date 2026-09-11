// Package medsos wires the medsos feature module: repository → service → handler.
//
// The module owns 5 endpoints under /medsos/* for CRUD operations on
// social media links per anggota.
package medsos

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/medsos/handler"
	"slam-team-api/internal/modules/core/medsos/repository"
	"slam-team-api/internal/modules/core/medsos/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the medsos feature's handler and route dependencies.
type Module struct {
	h      *handler.MedsosHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewMedsosRepository(db)
	svc := service.NewMedsosService(repo, auditor)
	return &Module{
		h:      handler.NewMedsosHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the 5 medsos endpoints under /medsos.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/medsos", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("medsos.read"), m.h.List)
	g.GET("/anggota-tersedia", m.perm.Require("medsos.read"), m.h.AnggotaTersedia)
	g.GET("/:id", m.perm.Require("medsos.read"), m.h.Detail)
	g.POST("", m.perm.Require("medsos.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("medsos.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("medsos.delete"), m.h.Delete)
}
