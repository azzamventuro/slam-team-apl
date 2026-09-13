// Package penugasan wires participant assignment (jadwal_peserta):
// repository → service → handler. It is the first producer of notifikasi
// rows, so it takes the notifikasi service at Initialize.
package penugasan

import (
	"slam-team-api/internal/middleware"
	notifservice "slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/internal/modules/core/penugasan/handler"
	"slam-team-api/internal/modules/core/penugasan/repository"
	"slam-team-api/internal/modules/core/penugasan/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the penugasan handler and route dependencies.
type Module struct {
	h      *handler.PesertaHandler
	jwtMgr *jwt.Manager
	perm   *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, perm *middleware.PermGuard, auditor *audit.Writer, notif *notifservice.NotifikasiService) *Module {
	repo := repository.NewPesertaRepository(db)
	svc := service.NewPesertaService(repo, notif, auditor)
	return &Module{
		h:      handler.NewPesertaHandler(svc),
		jwtMgr: jwtMgr,
		perm:   perm,
	}
}

// SetupRoutes mounts the assignment endpoints under the jadwal path
// (jadwal.assign = SA/Admin/Mod, jadwal.read for the list) and the
// assignee's own response under /peserta (bearer only — own-scope is
// enforced in the service via claims.AnggotaID).
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/jadwal", middleware.JWTAuth(m.jwtMgr))
	g.GET("/:id/peserta", m.perm.Require("jadwal.read"), m.h.List)
	g.POST("/:id/peserta", m.perm.Require("jadwal.assign"), m.h.Assign)
	g.POST("/:id/peserta/bulk", m.perm.Require("jadwal.assign"), m.h.BulkAssign)
	g.DELETE("/:id/peserta/:anggotaId", m.perm.Require("jadwal.assign"), m.h.Remove)

	p := rg.Group("/peserta", middleware.JWTAuth(m.jwtMgr))
	p.PATCH("/:id/respon", m.h.Respon)
}
