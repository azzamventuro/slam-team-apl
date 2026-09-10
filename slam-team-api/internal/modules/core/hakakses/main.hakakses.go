// Package hakakses wires the dynamic RBAC module: repository → service → handler.
//
// The module owns 7 endpoints under /hakakses/* and provides the PermGuard
// that other modules use for route-level permission checks.
package hakakses

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/hakakses/handler"
	"slam-team-api/internal/modules/core/hakakses/repository"
	"slam-team-api/internal/modules/core/hakakses/service"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Module bundles the hakakses feature's handler and route dependencies.
type Module struct {
	h         *handler.HakaksesHandler
	repo      *repository.HakaksesRepository
	jwtMgr    *jwt.Manager
	permGuard *middleware.PermGuard
}

// Initialize composes repository → service → handler and returns the Module.
// The permGuard is shared across all modules — it is created once in
// router.Setup and passed here.
func Initialize(db *gorm.DB, jwtMgr *jwt.Manager, permGuard *middleware.PermGuard, auditor *audit.Writer) *Module {
	repo := repository.NewHakaksesRepository(db)
	svc := service.NewHakaksesService(repo, permGuard, auditor)
	return &Module{
		h:         handler.NewHakaksesHandler(svc),
		repo:      repo,
		jwtMgr:    jwtMgr,
		permGuard: permGuard,
	}
}

// PermGuard exposes the shared PermGuard so other modules (e.g. auth) can
// call Effective for the GET /me/permissions endpoint.
func (m *Module) PermGuard() *middleware.PermGuard {
	return m.permGuard
}

// Repository exposes the hakakses repository so other modules (e.g. auth)
// can query GetUserPrimaryRole and GetPermVersion during login.
func (m *Module) Repository() *repository.HakaksesRepository {
	return m.repo
}

// SetupRoutes mounts the 7 RBAC endpoints under /hakakses.
//
// All endpoints require JWTAuth + RequireAdmin (level ≤ 10 or super admin).
// The dynamic matrix (PermGuard.Require) will be used by downstream modules
// that gate on per-permission checks.
func (m *Module) SetupRoutes(rg *gin.RouterGroup) {
	g := rg.Group("/hakakses")
	g.Use(middleware.JWTAuth(m.jwtMgr))
	g.Use(middleware.RequireAdmin()) // only Admin or Super Admin

	// Role CRUD
	g.POST("", m.h.CreateRole)
	g.GET("", m.h.ListRoles)
	g.GET("/:id", m.h.GetRole)
	g.PATCH("/:id", m.h.UpdateRole)
	g.DELETE("/:id", m.h.DeleteRole)

	// Permission matrix
	g.PUT("/:id/permissions", m.h.SetPermissions)

	// Permission catalogue (read-only, for the admin UI)
	g.GET("/permissions", m.h.ListPermissions)
}
