// Package pengaturan is the typed-settings feature module (mst_pengaturan).
//
// Settings carry enough metadata — label, tipe_nilai, opsi, satuan, keterangan,
// urutan — for the frontend to generate its own settings page, so adding a
// setting is one seeded row and no code on either side.
//
// The module is deliberately NOT gated by the dynamic RBAC guard: pengaturan is
// not one of the 19 mst_modul rows, so no pengaturan.* permission is ever
// seeded and there is nothing for the guard to look up. Access is decided by
// role instead (middleware.RequireAdmin), with the finer is_terkunci rule —
// super admin only — enforced in the service, where the affected rows are known.
package pengaturan

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/pengaturan/handler"
	"slam-team-api/internal/modules/core/pengaturan/repository"
	"slam-team-api/internal/modules/core/pengaturan/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the feature's handler and its route dependencies.
type Module struct {
	h      *handler.PengaturanHandler
	jwtMgr *jwt.Manager
}

// Initialize composes repository → service → handler.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, auditor *audit.Writer) *Module {
	repo := repository.NewPengaturanRepository(db)
	svc := service.NewPengaturanService(repo, auditor)
	return &Module{h: handler.NewPengaturanHandler(svc), jwtMgr: jwtMgr}
}

// SetupRoutes mounts the module's endpoints under /api/v1.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	// Unauthenticated: the app shell needs the club name and default timezone
	// before anyone has logged in. Returns is_publik rows only.
	rg.GET("/public/pengaturan", m.h.Public)

	g := rg.Group("/pengaturan", middleware.JWTAuth(m.jwtMgr), middleware.RequireAdmin())
	g.GET("", m.h.List)
	g.GET("/:grup", m.h.ListByGrup)
	g.PUT("", m.h.Update)
}
