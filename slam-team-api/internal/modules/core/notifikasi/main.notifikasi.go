// Package notifikasi wires the in-app inbox: repository → service → handler.
//
// Besides its four bearer-only endpoints under /notifikasi, the module exports
// its Service so other modules can fan messages out (penugasan now; izin,
// absensi, kta later). Register it in the router BEFORE any producer and pass
// Service() into that producer's Initialize.
package notifikasi

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/notifikasi/handler"
	"slam-team-api/internal/modules/core/notifikasi/repository"
	"slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the notifikasi handler, the reusable service, and the
// route dependencies.
type Module struct {
	h      *handler.NotifikasiHandler
	svc    *service.NotifikasiService
	jwtMgr *jwt.Manager
}

// Initialize composes repository → service → handler and returns the Module.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager) *Module {
	repo := repository.NewNotifikasiRepository(db)
	svc := service.NewNotifikasiService(repo)
	return &Module{
		h:      handler.NewNotifikasiHandler(svc),
		svc:    svc,
		jwtMgr: jwtMgr,
	}
}

// Service exposes the producer API for other modules' fan-out.
func (m *Module) Service() *service.NotifikasiService { return m.svc }

// SetupRoutes mounts the inbox endpoints. They are bearer-only: the inbox
// is own-scope, so there is no permission string to check (api-endpoints §4).
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/notifikasi", middleware.JWTAuth(m.jwtMgr))
	g.GET("", m.h.List)
	g.GET("/jumlah-belum-dibaca", m.h.JumlahBelumDibaca)
	// The literal segment is registered before the :id pattern so
	// "baca-semua" never parses as an id.
	g.PATCH("/baca-semua", m.h.BacaSemua)
	g.PATCH("/:id/baca", m.h.Baca)
}
