// Package handler exposes the participant-assignment endpoints under
// /jadwal/:id/peserta* and the assignee's own response under /peserta/:id/respon.
package handler

import (
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/penugasan/dto"
	"slam-team-api/internal/modules/core/penugasan/service"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// PesertaHandler holds the Gin handlers for the penugasan module.
type PesertaHandler struct {
	svc *service.PesertaService
}

// NewPesertaHandler creates a handler bound to the service.
func NewPesertaHandler(svc *service.PesertaService) *PesertaHandler {
	return &PesertaHandler{svc: svc}
}

// Assign handles POST /jadwal/:id/peserta.
func (h *PesertaHandler) Assign(c *gin.Context) {
	jadwalID, ok := parseParam(c, "id")
	if !ok {
		return
	}
	var req dto.AssignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	row, err := h.svc.Assign(c.Request.Context(), jadwalID, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, row)
}

// BulkAssign handles POST /jadwal/:id/peserta/bulk → {ditugaskan, dilewati}.
func (h *PesertaHandler) BulkAssign(c *gin.Context) {
	jadwalID, ok := parseParam(c, "id")
	if !ok {
		return
	}
	var req dto.BulkAssignReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	resp, err := h.svc.BulkAssign(c.Request.Context(), jadwalID, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// List handles GET /jadwal/:id/peserta.
func (h *PesertaHandler) List(c *gin.Context) {
	jadwalID, ok := parseParam(c, "id")
	if !ok {
		return
	}
	var query dto.ListPesertaQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	items, err := h.svc.List(c.Request.Context(), jadwalID, query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, items)
}

// Remove handles DELETE /jadwal/:id/peserta/:anggotaId — soft delete.
func (h *PesertaHandler) Remove(c *gin.Context) {
	jadwalID, ok := parseParam(c, "id")
	if !ok {
		return
	}
	anggotaID, ok := parseParam(c, "anggotaId")
	if !ok {
		return
	}
	if err := h.svc.Remove(c.Request.Context(), jadwalID, anggotaID, actorFromCtx(c)); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, nil)
}

// Respon handles PATCH /peserta/:id/respon — bearer only; the service
// enforces that the caller is the assignee.
func (h *PesertaHandler) Respon(c *gin.Context) {
	id, ok := parseParam(c, "id")
	if !ok {
		return
	}
	var req dto.ResponReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	row, err := h.svc.Respon(c.Request.Context(), id, req, actorFromCtx(c))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, row)
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
