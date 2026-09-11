// Package handler exposes the anggota (Master Data Anggota) HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/anggota/dto"
	"slam-team-api/internal/modules/core/anggota/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// AnggotaHandler holds the Gin handlers for anggota CRUD.
type AnggotaHandler struct {
	svc *service.AnggotaService
}

// NewAnggotaHandler creates a handler bound to the service.
func NewAnggotaHandler(svc *service.AnggotaService) *AnggotaHandler {
	return &AnggotaHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /anggota — paginated, filterable, sortable.
func (h *AnggotaHandler) List(c *gin.Context) {
	var query dto.ListAnggotaQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

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

	response.OK(c, pagination.Paginated[dto.AnggotaResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /anggota/:id — single anggota with joins.
func (h *AnggotaHandler) Detail(c *gin.Context) {
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

// Create handles POST /anggota — insert a new anggota row.
func (h *AnggotaHandler) Create(c *gin.Context) {
	var req dto.CreateAnggotaReq
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

// Update handles PUT /anggota/:id — full-replace update.
func (h *AnggotaHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateAnggotaReq
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

// Delete handles DELETE /anggota/:id — soft-delete with guard checks.
func (h *AnggotaHandler) Delete(c *gin.Context) {
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
