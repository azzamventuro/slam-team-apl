// Package handler exposes the prestasi (achievement / competition record)
// HTTP endpoints. The handler resolves scope from middleware context and
// delegates all business logic to the service layer.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/prestasi/dto"
	"slam-team-api/internal/modules/core/prestasi/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// PrestasiHandler holds the Gin handlers for prestasi CRUD.
type PrestasiHandler struct {
	svc *service.PrestasiService
}

// NewPrestasiHandler creates a handler bound to the service.
func NewPrestasiHandler(svc *service.PrestasiService) *PrestasiHandler {
	return &PrestasiHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /prestasi — paginated, filterable, sortable.
func (h *PrestasiHandler) List(c *gin.Context) {
	var query dto.ListPrestasiQuery
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
		response.FromError(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(query.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.PrestasiResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /prestasi/:id — single prestasi with joins.
func (h *PrestasiHandler) Detail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	resp, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// Create handles POST /prestasi — insert a new prestasi row.
func (h *PrestasiHandler) Create(c *gin.Context) {
	var req dto.CreatePrestasiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Create(c.Request.Context(), req, actor)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /prestasi/:id — full-replace update.
func (h *PrestasiHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdatePrestasiReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Update(c.Request.Context(), id, req, actor)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /prestasi/:id — soft-delete.
func (h *PrestasiHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	actor := actorFromCtx(c)
	if err := h.svc.SoftDelete(c.Request.Context(), id, actor); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, nil)
}

// PublicList handles GET /public/prestasi — public-safe, no auth.
func (h *PrestasiHandler) PublicList(c *gin.Context) {
	var query dto.ListPrestasiQuery
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

	items, total, err := h.svc.PublicList(c.Request.Context(), query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(query.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.PublicPrestasiItem]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// AnggotaTersedia handles GET /prestasi/anggota-tersedia — dropdown data.
func (h *PrestasiHandler) AnggotaTersedia(c *gin.Context) {
	items, err := h.svc.AnggotaTersedia(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, items)
}

// ── Helpers ──

func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

func actorFromCtx(c *gin.Context) service.Actor {
	cl := middleware.Claims(c)
	actor := service.Actor{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if cl != nil {
		actor.UserID = int64(cl.UserID)
	}
	return actor
}
