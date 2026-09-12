package inorga

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/inorga/handler"
	"slam-team-api/internal/modules/core/inorga/repository"
	"slam-team-api/internal/modules/core/inorga/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"
)

// Module wires inorga sub-packages and exposes route registration.
type Module struct {
	h      *handler.InorgaHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize builds the inorga dependency graph.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, aw *audit.Writer) *Module {
	repo := repository.NewInorgaRepository(db)
	svc := service.NewInorgaService(repo, aw)
	return &Module{h: handler.NewInorgaHandler(svc), jwtMgr: jwtMgr, perm: perm}
}

// SetupRoutes mounts CRUD under /inorga with JWT + permission guards.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/inorga", middleware.JWTAuth(m.jwtMgr))
	g.GET("",    m.perm.Require("inorga.read"),   m.h.List)
	g.GET("/:id", m.perm.Require("inorga.read"),  m.h.Detail)
	g.POST("",   m.perm.Require("inorga.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("inorga.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("inorga.delete"), m.h.Delete)
}
