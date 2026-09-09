// Package file is the centralised upload layer (mst_file + mst_file_varian),
// built once in Fase 0 before any module touches a file.
//
// Every other module stores an image or document as a `*_file_id bigint`
// pointing at mst_file.id — never as a path column of its own. Upload, EXIF
// auto-orientation, resizing into three variants and permission-checked serving
// are therefore written once here instead of being copied into eighteen
// modules.
//
// It is NOT one of the 19 mst_modul sidebar entries; it is infrastructure. It
// still owns two permissions (file.create, file.delete), because uploading and
// deleting are guarded actions like any other — see the seam note on
// SetupRoutes.
package file

import (
	"context"
	"strconv"
	"strings"

	"slam-team-api/internal/config"
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/file/handler"
	"slam-team-api/internal/modules/core/file/repository"
	"slam-team-api/internal/modules/core/file/service"
	pengaturanrepo "slam-team-api/internal/modules/core/pengaturan/repository"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"
	"slam-team-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Module bundles the feature's handler and its route dependencies.
type Module struct {
	h      *handler.FileHandler
	jwtMgr *jwt.Manager
}

// Initialize composes repository → service → handler.
//
// It returns an error rather than panicking on a bad STORAGE_ROOT: an
// unwritable storage directory is a configuration mistake the operator must see
// at boot, not an upload that reports success while the bytes go nowhere
// (rancangan Bab 6.6).
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, cfg *config.Config, auditor *audit.Writer) (*Module, error) {
	store, err := service.NewStorage(cfg.File.StorageRoot)
	if err != nil {
		return nil, err
	}
	logger.Info("file: storage siap",
		zap.String("root", store.Root()),
		zap.Int("maks_unggah_mb", cfg.File.MaxUploadMB))

	repo := repository.NewFileRepository(db)
	settings := &pengaturanSettings{repo: pengaturanrepo.NewPengaturanRepository(db)}
	svc := service.NewFileService(repo, store, settings, auditor, cfg.File.MaxUploadMB)
	return &Module{h: handler.NewFileHandler(svc), jwtMgr: jwtMgr}, nil
}

// SetupRoutes mounts the module's endpoints under /api/v1.
//
// SEAM for 05-hak-akses: file.create belongs to every signed-in role, so
// JWTAuth alone is already the correct gate in Fase 0; file.delete's ownership
// rule (Mod/User may only delete their own) is enforced in the service, where
// the file's created_by is known. When the dynamic guard lands, wrap POST with
// perm.Require("file.create") and DELETE with perm.Require("file.delete") —
// the service does not change. `file` is a permission namespace even though it
// is not an mst_modul row, so the RBAC seeder must still create both rows.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/files")

	g.POST("", middleware.JWTAuth(m.jwtMgr), m.h.Upload)
	g.DELETE("/:uuid", middleware.JWTAuth(m.jwtMgr), m.h.Delete)

	// Optional auth, not JWTAuth: one route serves both public artwork (an
	// <img> with no token) and private evidence (401/403 without the right
	// caller). Which one a uuid names is only knowable after the row is read,
	// so the decision belongs to the handler, not to a route-level gate.
	read := g.Group("", middleware.JWTOptional(m.jwtMgr))
	read.GET("/:uuid", m.h.Detail)
	read.GET("/:uuid/:varian", m.h.Serve)
}

// pengaturanSettings adapts the pengaturan module's repository to the narrow
// SettingsReader the file service declares, so the variant pixel sizes come
// from mst_pengaturan grup "file" instead of being hardcoded — change the
// setting, and the next upload follows.
//
// ponytail: reads the row per upload. Add a short cache here if the settings
// table ever shows up in an upload profile.
type pengaturanSettings struct {
	repo *pengaturanrepo.PengaturanRepository
}

// Int returns the setting as an integer, falling back to def for a missing,
// empty or unparsable row. A misconfigured setting must never fail an upload —
// it degrades to the documented default, loudly.
func (p *pengaturanSettings) Int(ctx context.Context, kunci string, def int) int {
	row, err := p.repo.FindByKunci(ctx, kunci)
	if err != nil {
		logger.Warn("file: pengaturan tidak terbaca, memakai nilai bawaan",
			zap.String("kunci", kunci), zap.Int("bawaan", def), zap.Error(err))
		return def
	}
	if row.Nilai == nil {
		return def
	}
	n, err := strconv.Atoi(strings.TrimSpace(*row.Nilai))
	if err != nil || n <= 0 {
		logger.Warn("file: pengaturan bukan bilangan positif, memakai nilai bawaan",
			zap.String("kunci", kunci), zap.String("nilai", *row.Nilai), zap.Int("bawaan", def))
		return def
	}
	return n
}
