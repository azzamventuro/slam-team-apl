// Package middleware — PermGuard is the dynamic RBAC permission guard. It
// replaces RequireAdmin for modules that participate in the matrix.
//
// PermGuard caches each role's effective permission set in memory (sync.Map).
// The cache is keyed by role_id. When the matrix changes (role_permission
// insert/update/delete), the service calls Invalidate(roleID) so the next
// request re-fetches from DB.
package middleware

import (
	"context"
	"sync"

	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ── PermGuard ──

// PermGuard holds the per-role permission cache and the DB handle to
// re-populate it. Create once in router.Setup, pass to all modules.
type PermGuard struct {
	db    *gorm.DB
	cache sync.Map // key: int64 (role_id), value: *permCacheEntry
}

type permCacheEntry struct {
	set map[string]string // perm_kode → cakupan ("semua" | "milik_sendiri" | ...)
}

// NewPermGuard creates a PermGuard. The db handle must be the same GORM
// instance used by the rest of the application.
func NewPermGuard(db *gorm.DB) *PermGuard {
	return &PermGuard{db: db}
}

// Require returns a Gin middleware that checks whether the authenticated user's
// role holds the named permission (e.g. "anggota.create"). Super Admin bypasses
// every check. The permission must be an exact mst_permission.kode value.
//
// Require must be chained AFTER JWTAuth (it reads claims from context).
func (pg *PermGuard) Require(kodePerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c) // from context.go
		if cl == nil {
			response.Unauthorized(c, "tidak terautentikasi")
			c.Abort()
			return
		}
		// Super Admin bypass.
		if cl.IsSuper {
			c.Next()
			return
		}
		// Level 0 without IsSuper is a broken token — deny.
		if cl.RoleLevel == 0 {
			response.Forbidden(c, "token tanpa peran valid")
			c.Abort()
			return
		}
		// Look up effective permissions from cache/DB.
		perms, err := pg.Effective(context.Background(), cl.RoleID)
		if err != nil {
			logger.Error("permguard: gagal load permissions", zap.Error(err))
			response.Internal(c, err)
			c.Abort()
			return
		}
		cakupan, ok := perms[kodePerm]
		if !ok {
			response.Forbidden(c, "tidak diizinkan: "+kodePerm)
			c.Abort()
			return
		}
		// Store cakupan in context so downstream handlers can branch on
		// "semua" vs "milik_sendiri".
		c.Set("cakupan_"+kodePerm, cakupan)
		c.Next()
	}
}

// RequireAny returns a middleware that grants access if the user holds ANY of
// the listed permissions. Used for endpoints like GET /hakakses/roles where
// different roles can access via different permissions.
func (pg *PermGuard) RequireAny(kodePerms ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil {
			response.Unauthorized(c, "tidak terautentikasi")
			c.Abort()
			return
		}
		if cl.IsSuper || (cl.RoleLevel > 0 && cl.RoleLevel <= 10) {
			c.Next()
			return
		}
		perms, err := pg.Effective(context.Background(), cl.RoleID)
		if err != nil {
			response.Internal(c, err)
			c.Abort()
			return
		}
		for _, kode := range kodePerms {
			if _, ok := perms[kode]; ok {
				cakupan := perms[kode]
				c.Set("cakupan_"+kode, cakupan)
				c.Next()
				return
			}
		}
		response.Forbidden(c, "tidak diizinkan")
		c.Abort()
	}
}

// Effective returns the permission set for a role, reading from cache first.
// Returns a map of perm_kode → cakupan.
func (pg *PermGuard) Effective(ctx context.Context, roleID int64) (map[string]string, error) {
	if v, ok := pg.cache.Load(roleID); ok {
		entry := v.(*permCacheEntry)
		// shallow copy to avoid races on the map
		out := make(map[string]string, len(entry.set))
		for k, v2 := range entry.set {
			out[k] = v2
		}
		return out, nil
	}
	return pg.reload(ctx, roleID)
}

// reload fetches the effective permissions from DB and stores them in cache.
func (pg *PermGuard) reload(ctx context.Context, roleID int64) (map[string]string, error) {
	type row struct {
		Kode    string `gorm:"column:kode"`
		Cakupan string `gorm:"column:cakupan"`
	}
	var rows []row
	err := pg.db.WithContext(ctx).
		Table("role_permission rp").
		Joins("JOIN mst_permission p ON p.id = rp.permission_id").
		Joins("JOIN mst_modul m ON m.id = p.modul_id AND m.is_aktif = true").
		Where("rp.role_id = ?", roleID).
		Select("p.kode, rp.cakupan::text AS cakupan").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	set := make(map[string]string, len(rows))
	for _, r := range rows {
		set[r.Kode] = r.Cakupan
	}
	pg.cache.Store(roleID, &permCacheEntry{set: set})
	out := make(map[string]string, len(set))
	for k, v := range set {
		out[k] = v
	}
	return out, nil
}

// Invalidate removes the cached permission set for a role so the next request
// re-fetches from DB. Called by the service after any matrix change.
func (pg *PermGuard) Invalidate(roleID int64) {
	pg.cache.Delete(roleID)
}

// InvalidateAll clears the entire cache (used after bulk operations).
func (pg *PermGuard) InvalidateAll() {
	pg.cache.Clear()
}

// ── RequireSelf helper ──

// RequireSelf returns a middleware that checks cakupan of the given permission.
// If cakupan == "milik_sendiri", the user must own the resource being accessed.
// The ownership check compares the user_id JWT claim with the :user_id path param.
func RequireSelf(kodePerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cakupan := GetCakupan(c, kodePerm)
		if cakupan == "" || cakupan == "semua" {
			c.Next()
			return
		}
		// cakupan == "milik_sendiri" → the request must target the caller's own resource.
		cl := Claims(c)
		if cl == nil {
			response.Unauthorized(c, "tidak terautentikasi")
			c.Abort()
			return
		}
		targetID := c.Param("user_id")
		if targetID == "" {
			targetID = c.Param("id")
		}
		if targetID != "" {
			var parsed int64
			for _, ch := range targetID {
				if ch >= '0' && ch <= '9' {
					parsed = parsed*10 + int64(ch-'0')
				} else {
					parsed = -1
					break
				}
			}
			if parsed > 0 && parsed != int64(cl.UserID) {
				response.Forbidden(c, "hanya dapat mengakses data sendiri")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// GetCakupan retrieves the cakupan value stored by a Require middleware for
// the given permission kode. Returns "" if not set.
func GetCakupan(c *gin.Context, kodePerm string) string {
	v, _ := c.Get("cakupan_" + kodePerm)
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}
