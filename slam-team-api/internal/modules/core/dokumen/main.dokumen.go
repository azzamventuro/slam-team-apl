// Package dokumen wires the dokumen feature module: repository → service → handler.
//
// The module owns 5 endpoints under /dokumen/* for CRUD operations on
// polymorphic document attachments. It depends on the file layer to resolve
// file_uuid → file_id (mst_file.id).
package dokumen

import (
	"slam-team-api/internal/middleware"
	filerepository "slam-team-api/internal/modules/core/file/repository"
	"slam-team-api/internal/modules/core/dokumen/handler"
	"slam-team-api/internal/modules/core/dokumen/repository"
	"slam-team-api/internal/modules/core/dokumen/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the dokumen feature's handler and route dependencies.
type Module struct {
	h      *handler.DokumenHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
// fileRepo is the shared file repository used to resolve file_uuid → file_id.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer, fileRepo *filerepository.FileRepository) *Module {
	repo := repository.NewDokumenRepository(db)
	svc := service.NewDokumenService(repo, fileRepo, auditor)
	return &Module{
		h:      handler.NewDokumenHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the 5 dokumen endpoints under /dokumen.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/dokumen", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("dokumen.read"), m.h.List)
	g.GET("/:id", m.perm.Require("dokumen.read"), m.h.Detail)
	g.POST("", m.perm.Require("dokumen.create"), m.h.Create)
	g.PUT("/:id", m.perm.Require("dokumen.update"), m.h.Update)
	g.DELETE("/:id", m.perm.Require("dokumen.delete"), m.h.Delete)
}
