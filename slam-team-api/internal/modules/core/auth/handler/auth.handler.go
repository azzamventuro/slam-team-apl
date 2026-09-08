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
	res, err := h.svc.Login(req)
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
		"id":    claims.UserID,
		"email": claims.Email,
		"name":  claims.Name,
	})
}
