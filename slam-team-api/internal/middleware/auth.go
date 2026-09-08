package middleware

import (
	"strings"

	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth validates the Authorization: Bearer <token> header and stores the
// decoded claims in the request context (retrieve with middleware.Claims).
func JWTAuth(m *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing bearer token")
			c.Abort()
			return
		}
		claims, err := m.Verify(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}
