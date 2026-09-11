// Package handler exposes the dokumen (polymorphic document attachments) HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/dokumen/dto"
	"slam-team-api/internal/modules/core/dokumen/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// DokumenHandler holds the Gin handlers for dokumen CRUD.
type DokumenHandler struct {
	svc *service.DokumenService
}

// NewDokumenHandler creates a handler bound to the service.
func NewDokumenHandler(svc *service.DokumenService) *DokumenHandler {
	return &DokumenHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /dokumen — paginated, filterable, sortable.
func (h *DokumenHandler) List(c *gin.Context) {
	var query dto.ListDokumenQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	query.Normalize()

	items, total, err := h.svc.List(c.Request.Context(), query)
	if err != nil {
		respondError(c, err)
		return
	}

	// Recalculate pagination after service normalization.
	perPage := query.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.DokumenResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  perPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /dokumen/:id — single document.
func (h *DokumenHandler) Detail(c *gin.Context) {
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

// Create handles POST /dokumen — insert a new document attachment.
func (h *DokumenHandler) Create(c *gin.Context) {
	var req dto.CreateDokumenReq
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

// Update handles PUT /dokumen/:id — update an existing document attachment.
func (h *DokumenHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateDokumenReq
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

// Delete handles DELETE /dokumen/:id — soft-delete.
func (h *DokumenHandler) Delete(c *gin.Context) {
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
	var userID int64
	if cl != nil {
		userID = int64(cl.UserID)
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
