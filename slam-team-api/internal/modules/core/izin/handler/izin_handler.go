// Package handler exposes the izin endpoints under /izin. The handler
// resolves the caller and the izin.read cakupan from the middleware context
// and delegates every rule to the service.
package handler

import (
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/izin/dto"
	"slam-team-api/internal/modules/core/izin/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// IzinHandler holds the Gin handlers for the izin module.
type IzinHandler struct {
	svc *service.IzinService
}

// NewIzinHandler creates a handler bound to the service.
func NewIzinHandler(svc *service.IzinService) *IzinHandler {
	return &IzinHandler{svc: svc}
}

// Create handles POST /izin.
func (h *IzinHandler) Create(c *gin.Context) {
	var req dto.CreateIzinReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	scope := middleware.GetCakupan(c, "izin.create")
	resp, err := h.svc.Create(c.Request.Context(), req, scope, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, resp)
}

// List handles GET /izin — paginated; owner filter from the izin.read cakupan.
func (h *IzinHandler) List(c *gin.Context) {
	var query dto.ListIzinQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	scope := middleware.GetCakupan(c, "izin.read")
	items, total, err := h.svc.List(c.Request.Context(), query, scope, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, pagination.New(items, query.ListQuery, total))
}

// Approve handles PATCH /izin/:id/approve.
func (h *IzinHandler) Approve(c *gin.Context) {
	id, ok := parseParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.svc.Approve(c.Request.Context(), id, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// Tolak handles PATCH /izin/:id/tolak.
func (h *IzinHandler) Tolak(c *gin.Context) {
	id, ok := parseParam(c, "id")
	if !ok {
		return
	}
	var req dto.TolakReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	resp, err := h.svc.Tolak(c.Request.Context(), id, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// ── Helpers ──

// parseParam reads a positive int64 path param; on failure it has already
// written the 400 and returns ok=false.
func parseParam(c *gin.Context, name string) (int64, bool) {
	v, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || v <= 0 {
		response.BadRequest(c, name+" tidak valid", nil)
		return 0, false
	}
	return v, true
}

func actorFromCtx(c *gin.Context) service.Actor {
	cl := middleware.Claims(c)
	a := service.Actor{IPAddress: c.ClientIP(), UserAgent: c.Request.UserAgent()}
	if cl != nil {
		id := int64(cl.UserID)
		a.UserID = &id
		a.AnggotaID = cl.AnggotaID
	}
	return a
}
