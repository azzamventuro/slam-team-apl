// Package handler exposes the instansi (Master Data Instansi / Sekolah) HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/instansi/dto"
	"slam-team-api/internal/modules/core/instansi/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// InstansiHandler holds the Gin handlers for instansi CRUD.
type InstansiHandler struct {
	svc *service.InstansiService
}

// NewInstansiHandler creates a handler bound to the service.
func NewInstansiHandler(svc *service.InstansiService) *InstansiHandler {
	return &InstansiHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /instansi — paginated, filterable, sortable.
func (h *InstansiHandler) List(c *gin.Context) {
	var query dto.ListInstansiQuery
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

	response.OK(c, pagination.Paginated[dto.InstansiResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /instansi/:id — single instansi with joins.
func (h *InstansiHandler) Detail(c *gin.Context) {
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

// Create handles POST /instansi — insert a new instansi row.
func (h *InstansiHandler) Create(c *gin.Context) {
	var req dto.CreateInstansiReq
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

// Update handles PUT /instansi/:id — full-replace update.
func (h *InstansiHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateInstansiReq
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

// Delete handles DELETE /instansi/:id — soft-delete with anggota guard.
func (h *InstansiHandler) Delete(c *gin.Context) {
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
