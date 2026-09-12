// Package handler exposes the hakakses (RBAC) HTTP endpoints.
package handler

import (
	"errors"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/hakakses/dto"
	"slam-team-api/internal/modules/core/hakakses/service"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type HakaksesHandler struct {
	svc *service.HakaksesService
}

func NewHakaksesHandler(svc *service.HakaksesService) *HakaksesHandler {
	return &HakaksesHandler{svc: svc}
}

// ── Role CRUD ──

// POST /hakakses/roles
func (h *HakaksesHandler) CreateRole(c *gin.Context) {
	var req dto.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "validasi gagal", err.Error())
		return
	}
	actor := actorFromCtx(c)
	role, err := h.svc.CreateRole(c.Request.Context(), req, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Created(c, role)
}

// GET /hakakses/roles
func (h *HakaksesHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, roles)
}

// GET /hakakses/roles/:id
func (h *HakaksesHandler) GetRole(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	role, err := h.svc.GetRoleDetail(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, role)
}

// PATCH /hakakses/roles/:id
func (h *HakaksesHandler) UpdateRole(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	var req dto.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "validasi gagal", err.Error())
		return
	}
	actor := actorFromCtx(c)
	role, err := h.svc.UpdateRole(c.Request.Context(), id, req, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, role)
}

// DELETE /hakakses/roles/:id
func (h *HakaksesHandler) DeleteRole(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	actor := actorFromCtx(c)
	if err := h.svc.DeleteRole(c.Request.Context(), id, actor); err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, nil)
}

// ── Permission matrix ──

// PUT /hakakses/roles/:id/permissions
func (h *HakaksesHandler) SetPermissions(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	var req dto.SetPermissionsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "validasi gagal", err.Error())
		return
	}
	actor := actorFromCtx(c)
	if err := h.svc.SetRolePermissions(c.Request.Context(), id, req, actor); err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, nil)
}

// GET /hakakses/permissions
func (h *HakaksesHandler) ListPermissions(c *gin.Context) {
	moduls, err := h.svc.ListModulPermissions(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, moduls)
}

// ── Helpers ──

func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

func actorFromCtx(c *gin.Context) service.Actor {
	cl := middleware.Claims(c)
	var userID *int64
	if cl != nil {
		id := int64(cl.UserID)
		userID = &id
	}
	return service.Actor{
		UserID:    userID,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		response.NotFound(c, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		response.Forbidden(c, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		response.Unprocess(c, err.Error(), nil)
	case errors.Is(err, apperr.ErrValidation):
		response.BadRequest(c, err.Error(), nil)
	default:
		response.Internal(c, err)
	}
}
