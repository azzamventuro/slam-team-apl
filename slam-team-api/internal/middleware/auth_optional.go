package middleware

import (
	"strings"

	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTOptional decodes a bearer token when one is present and lets the request
// through when it is not. Read the result with middleware.Claims, which returns
// nil for an anonymous caller.
//
// It exists for routes that serve BOTH public and private material behind one
// URL — GET /files/{uuid}/{varian} is the case: a logo must load in an <img>
// tag with no token, while an absensi selfie behind the same route shape must
// not. Only the handler knows which the uuid names, so the 401 is raised there
// rather than here.
//
// An invalid or expired token is treated as anonymous rather than rejected: the
// caller is then refused by the same rule that refuses a missing one, and a
// stale token never blocks a genuinely public file. Use JWTAuth wherever a
// token is actually required.
func JWTOptional(m *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Next()
			return
		}
		claims, err := m.Verify(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.Next()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}
