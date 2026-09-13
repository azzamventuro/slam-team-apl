// Package router builds the Gin engine and mounts the /api/v1 route group.
package router

import (
	"net/http"
	"time"

	"slam-team-api/internal/config"
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/anggota"
	"slam-team-api/internal/modules/core/auth"
	"slam-team-api/internal/modules/core/dokumen"
	"slam-team-api/internal/modules/core/file"
	filerepository "slam-team-api/internal/modules/core/file/repository"
	"slam-team-api/internal/modules/core/hakakses"
	"slam-team-api/internal/modules/core/inorga"
	"slam-team-api/internal/modules/core/instansi"
	"slam-team-api/internal/modules/core/izin"
	"slam-team-api/internal/modules/core/jadwal"
	"slam-team-api/internal/modules/core/lokasi"
	"slam-team-api/internal/modules/core/medsos"
	"slam-team-api/internal/modules/core/notifikasi"
	"slam-team-api/internal/modules/core/pengaturan"
	"slam-team-api/internal/modules/core/penugasan"
	"slam-team-api/internal/modules/core/prestasi"
	"slam-team-api/internal/modules/core/profileclub"
	"slam-team-api/internal/modules/core/unit"
	"slam-team-api/internal/modules/core/usermgmt"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Setup builds the engine, attaches global middleware, exposes /health, and
// registers every feature module under /api/v1.
//
// It returns an error when a module cannot be constructed — the file layer
// refuses to start on an unwritable STORAGE_ROOT — so a misconfiguration stops
// the process at boot instead of surfacing as a failed upload later.
func Setup(cfg *config.Config, db *gorm.DB, jwtMgr *jwt.Manager, rdb *goredis.Client) (*gin.Engine, error) {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	middleware.Global(engine, cfg)

	// Liveness/readiness probe for load balancers and uptime monitors, not a
	// business endpoint — deliberately flat JSON, outside /api/v1 and outside
	// the {success,message,data} envelope.
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"time":    time.Now().UTC().Format(time.RFC3339),
			"version": cfg.App.Version,
			"env":     cfg.App.Env,
			"redis":   rdb != nil,
		})
	})

	apiV1 := engine.Group("/api/v1")

	// Shared audit writer: every module that changes state records a
	// log_aktivitas row through it.
	auditor := audit.NewWriter(db)

	// Shared PermGuard: dynamic RBAC permission cache. Created once, passed
	// to every module that needs route-level permission checks.
	permGuard := middleware.NewPermGuard(db)

	// --- Register feature modules here ---

	// Hakakses (RBAC): role CRUD, permission matrix, permission catalogue.
	// Registered FIRST because auth depends on hakakses repo for role claims
	// in login tokens and the /me/permissions endpoint.
	hakaksesModule := hakakses.Initialize(db, jwtMgr, permGuard, auditor)
	hakaksesModule.SetupRoutes(apiV1)

	// Auth: depends on hakakses repo (for GetUserPrimaryRole / GetPermVersion)
	// and permGuard (for Effective in /me/permissions).
	auth.Initialize(db, jwtMgr, hakaksesModule.Repository(), permGuard, auditor, cfg.JWT.RefreshTTL).SetupRoutes(apiV1)

	pengaturan.Initialize(db, jwtMgr, auditor).SetupRoutes(apiV1)

	// The file layer reads its variant sizes from mst_pengaturan, so it is
	// registered after it.
	fileModule, err := file.Initialize(db, jwtMgr, cfg, auditor)
	if err != nil {
		return nil, err
	}
	fileModule.SetupRoutes(apiV1)

	// Instansi (Master Data Instansi / Sekolah): CRUD with RBAC permission guards.
	// Registered after file (logo uploads) and hakakses (permission checks).
	instansi.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Lokasi (Master Data Lokasi / geofence + timezone): CRUD with RBAC permission
	// guards. Registered after file (foto uploads). Delete is guarded by the
	// jadwal usage count (409 while any active jadwal references the lokasi).
	lokasi.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Jadwal (schedule definitions + generated jadwal_sesi): CRUD, generate-sesi,
	// batalkan sesi. Registered after lokasi (geofence snapshot) and inorga (FK).
	jadwal.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Notifikasi (in-app inbox, own-scope, bearer only). Registered before
	// every producer: its Service() is the only sanctioned writer of
	// notifikasi rows, handed to penugasan now and izin/absensi/kta later.
	notifikasiModule := notifikasi.Initialize(db, jwtMgr)
	notifikasiModule.SetupRoutes(apiV1)

	// Penugasan (jadwal_peserta): assign/bulk/list/remove under /jadwal/:id/peserta*
	// (jadwal.assign / jadwal.read) + the assignee's own PATCH /peserta/:id/respon.
	// Registered after jadwal (FK) and notifikasi (fan-out on assign).
	penugasan.Initialize(db, jwtMgr, permGuard, auditor, notifikasiModule.Service()).SetupRoutes(apiV1)

	// Izin (absensi_izin): submit / list (cakupan) / approve / tolak under
	// /izin. Registered after jadwal (sesi FK) and notifikasi (fan-out on
	// decision). The absensi linkage (absensi.izin_id) is owned by 18-absensi.
	izin.Initialize(db, jwtMgr, permGuard, auditor, notifikasiModule.Service()).SetupRoutes(apiV1)

	// Anggota (Master Data Anggota / Core Member Record): CRUD with RBAC permission guards.
	// Registered after instansi (FK validation) and file (foto uploads).
	anggota.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Unit (Master Data Unit / Airsoft Gun Registry): CRUD with RBAC permission guards.
	// Registered after anggota (FK validation) and file (foto uploads).
	unit.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Prestasi (Achievement / Competition Records): CRUD with RBAC permission guards.
	// Registered after anggota (FK validation) and file (flyer/foto uploads).
	prestasiModule := prestasi.Initialize(db, jwtMgr, permGuard, auditor)
	prestasiModule.SetupRoutes(apiV1)

	// Inorga (Kepengurusan / Organisational Periods): CRUD with RBAC permission guards.
	// Registered after file (logo/banner/SK uploads).
	inorga.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Medsos (Social Media Links per Anggota): CRUD with RBAC permission guards.
	// Registered after anggota (FK validation).
	medsos.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// Dokumen (Polymorphic Document Attachments): CRUD with RBAC permission guards.
	// Registered after file (file_uuid → file_id resolution) and before usermgmt.
	// A separate FileRepository instance is created here since the file module
	// does not export its repository.
	dokumenFileRepo := filerepository.NewFileRepository(db)
	dokumen.Initialize(db, jwtMgr, permGuard, auditor, dokumenFileRepo).SetupRoutes(apiV1)

	// ProfileClub (Singleton Club Identity): CRUD with RBAC permission guards.
	// Registered after file (logo/banner uploads) and before usermgmt.
	profileClubModule := profileclub.Initialize(db, jwtMgr, permGuard, auditor)
	profileClubModule.SetupRoutes(apiV1)

	// Usermgmt (User Management surfaces: admin / moderator / user): CRUD over
	// users + user_role filtered by role-level band. Anti-escalation enforced.
	// Registered after anggota (FK validation) and hakakses (permission checks).
	usermgmt.Initialize(db, jwtMgr, permGuard, auditor).SetupRoutes(apiV1)

	// ── Public routes (no auth required) ──
	publicV1 := engine.Group("/api/public")
	prestasiModule.PublicSetupRoutes(publicV1)
	profileClubModule.PublicSetupRoutes(publicV1)

	return engine, nil
}
