// Package handler holds the auth module's Gin HTTP handlers.
package handler

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/dto"
	"slam-team-api/internal/modules/core/auth/service"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AuthHandler adapts HTTP requests to the auth service.
type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

// clientInfo lifts the request fingerprint out of Gin for sesi_login/audit.
func clientInfo(c *gin.Context) service.ClientInfo {
	return service.ClientInfo{IP: c.ClientIP(), UserAgent: c.Request.UserAgent()}
}

// Login handles POST /auth/login — username OR email + password.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req, clientInfo(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, res)
}

// Refresh handles POST /auth/refresh — rotates the refresh token.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), req, clientInfo(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, res)
}

// Logout handles POST /auth/logout — revokes the caller's session.
func (h *AuthHandler) Logout(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "not authenticated")
		return
	}
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	if err := h.svc.Logout(c.Request.Context(), claims, req, clientInfo(c)); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, nil)
}

// Me handles GET /auth/me — the account, its role, and the linked anggota.
func (h *AuthHandler) Me(c *gin.Context) {
	claims := middleware.Claims(c)
	if claims == nil {
		response.Unauthorized(c, "not authenticated")
		return
	}
	res, err := h.svc.Me(c.Request.Context(), claims)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, res)
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
