// Package handler holds the auth module's Gin HTTP handlers.
package handler

import (
	"errors"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/dto"
	"slam-team-api/internal/modules/core/auth/service"
	"slam-team-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

// AuthHandler adapts HTTP requests to the auth service.
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validation failed", err.Error())
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.Internal(c, err)
		return
	}
	response.OK(c, res)
}

// Me handles GET /auth/me — returns the authenticated user's claims.
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "not authenticated")
		return
	}
	response.OK(c, gin.H{
		"id":          claims.UserID,
		"email":       claims.Email,
		"name":        claims.Name,
		"role_id":     claims.RoleID,
		"role_level":  claims.RoleLevel,
		"is_super":    claims.IsSuper,
		"perm_version": claims.PermVersion,
	})
}

// MePermissions handles GET /auth/me/permissions — returns the user's effective
// permission set from the dynamic RBAC matrix.
func (h *AuthHandler) MePermissions(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "not authenticated")
		return
	}
	resp, err := h.svc.MyPermissions(c.Request.Context(), claims)
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, resp)
}
