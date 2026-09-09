package middleware

import (
	"slam-team-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// LevelAdmin is mst_role.level for Admin. Levels are ordered by privilege with
// 0 (Super Admin) the highest, so "Admin or better" is level <= LevelAdmin.
const LevelAdmin = 10

// RequireAdmin gates the system pages that have no seeded modul.aksi permission
// of their own — pengaturan is one: it is not among the 19 mst_modul rows, so
// the dynamic PermGuard has nothing to look up. Access is decided by role
// instead: super admin, or a role at Admin level or above.
//
// Level 0 alone is not accepted as "better than Admin": a token that carries no
// role claims at all decodes to RoleLevel 0, and that must be denied, not
// mistaken for the super admin (who is identified by IsSuper).
//
// Must be mounted after JWTAuth. Finer rules that depend on the row being
// touched — is_terkunci is super-admin-only — belong in the service, because a
// route-level gate cannot see which rows a request will change.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		cl := Claims(c)
		if cl == nil {
			response.Unauthorized(c, "tidak terautentikasi")
			c.Abort()
			return
		}
		if cl.IsSuper || (cl.RoleLevel > 0 && cl.RoleLevel <= LevelAdmin) {
			c.Next()
			return
		}
		response.Forbidden(c, "tidak diizinkan")
		c.Abort()
	}
}
