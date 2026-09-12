// Package handler exposes the lokasi (Master Data Lokasi) HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/lokasi/dto"
	"slam-team-api/internal/modules/core/lokasi/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// LokasiHandler holds the Gin handlers for lokasi CRUD.
type LokasiHandler struct {
	svc *service.LokasiService
}

// NewLokasiHandler creates a handler bound to the service.
func NewLokasiHandler(svc *service.LokasiService) *LokasiHandler {
	return &LokasiHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /lokasi — paginated, filterable, sortable.
func (h *LokasiHandler) List(c *gin.Context) {
	var query dto.ListLokasiQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	// Normalise pagination.
	if query.Page < 1 {
		query.Page = 1
	}
	if query.PerPage < 1 {
		query.PerPage = 20
	}
	if query.PerPage > 100 {
		query.PerPage = 100
	}

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		response.Internal(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(query.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.LokasiResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /lokasi/:id — single lokasi with joins.
func (h *LokasiHandler) Detail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	resp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Create handles POST /lokasi — insert a new lokasi row.
func (h *LokasiHandler) Create(c *gin.Context) {
	var req dto.CreateLokasiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Create(c.Request.Context(), req, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /lokasi/:id — full-replace update.
func (h *LokasiHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateLokasiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Update(c.Request.Context(), id, req, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /lokasi/:id — soft-delete.
func (h *LokasiHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	actor := actorFromCtx(c)
	if err := h.svc.SoftDelete(c.Request.Context(), id, actor); err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, nil)
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
	response.FromError(c, err)
}
