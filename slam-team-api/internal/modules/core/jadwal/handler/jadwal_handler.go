// Package handler exposes the jadwal HTTP endpoints: schedule CRUD, the
// session generator, the session list, and session cancellation.
package handler

import (
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/jadwal/dto"
	"slam-team-api/internal/modules/core/jadwal/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// JadwalHandler holds the Gin handlers for the jadwal module.
type JadwalHandler struct {
	svc *service.JadwalService
}

// NewJadwalHandler creates a handler bound to the service.
func NewJadwalHandler(svc *service.JadwalService) *JadwalHandler {
	return &JadwalHandler{svc: svc}
}

// ── Jadwal CRUD ──

// List handles GET /jadwal — paginated, filterable, sortable.
func (h *JadwalHandler) List(c *gin.Context) {
	var query dto.ListJadwalQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	lq := pagination.ListQuery{Page: query.Page, PerPage: query.PerPage, Q: query.Q, Sort: query.Sort}
	response.OK(c, pagination.New(items, lq, total))
}

// Detail handles GET /jadwal/:id.
func (h *JadwalHandler) Detail(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	resp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// Create handles POST /jadwal.
func (h *JadwalHandler) Create(c *gin.Context) {
	var req dto.CreateJadwalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	resp, err := h.svc.Create(c.Request.Context(), req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /jadwal/:id — full-replace update.
func (h *JadwalHandler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.UpdateJadwalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	resp, err := h.svc.Update(c.Request.Context(), id, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /jadwal/:id — soft delete.
func (h *JadwalHandler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.SoftDelete(c.Request.Context(), id, actorFromCtx(c)); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, nil)
}

// ── Sesi ──

// GenerateSesi handles POST /jadwal/:id/generate-sesi. An empty body is
// valid: every field defaults to the schedule's own values.
func (h *JadwalHandler) GenerateSesi(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.GenerateSesiReq
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Unprocess(c, "validasi gagal", validator.Explain(err))
			return
		}
	}
	resp, err := h.svc.GenerateSesi(c.Request.Context(), id, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// ListSesi handles GET /jadwal/:id/sesi.
func (h *JadwalHandler) ListSesi(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var query dto.ListSesiQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	items, err := h.svc.ListSesi(c.Request.Context(), id, query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, items)
}

// BatalkanSesi handles PATCH /sesi/:id/batalkan.
func (h *JadwalHandler) BatalkanSesi(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.BatalkanSesiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	resp, err := h.svc.BatalkanSesi(c.Request.Context(), id, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// ── Helpers ──

// parseID reads the :id path param; on failure it has already written the
// 400 and returns ok=false.
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "id tidak valid", nil)
		return 0, false
	}
	return id, true
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
