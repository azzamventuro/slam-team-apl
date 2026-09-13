// Package izin wires excuse requests (absensi_izin): repository → service →
// handler. It is a notifikasi producer (izin_disetujui / izin_ditolak), so it
// takes the notifikasi service at Initialize.
package izin

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/izin/handler"
	"slam-team-api/internal/modules/core/izin/repository"
	"slam-team-api/internal/modules/core/izin/service"
	notifservice "slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the izin handler and route dependencies.
type Module struct {
	h      *handler.IzinHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer, notif *notifservice.NotifikasiService) *Module {
	repo := repository.NewIzinRepository(db)
	svc := service.NewIzinService(repo, notif, auditor)
	return &Module{
		h:      handler.NewIzinHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the izin endpoints under /izin (api-endpoints §9):
// izin.create for every logged-in role, izin.read with cakupan (semua vs
// milik_sendiri, enforced in the service), izin.approve for both decisions.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/izin", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.perm.Require("izin.read"), m.h.List)
	g.POST("", m.perm.Require("izin.create"), m.h.Create)
	g.PATCH("/:id/approve", m.perm.Require("izin.approve"), m.h.Approve)
	g.PATCH("/:id/tolak", m.perm.Require("izin.approve"), m.h.Tolak)
}
