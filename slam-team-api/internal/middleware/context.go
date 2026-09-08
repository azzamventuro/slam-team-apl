package middleware

import (
	"slam-team-api/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const claimsKey = "auth_claims"

// Claims returns the authenticated user's JWT claims, or nil when the request
// did not pass through JWTAuth.
func Claims(c *gin.Context) *jwt.Claims {
	v, ok := c.Get(claimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*jwt.Claims)
	return claims
}
