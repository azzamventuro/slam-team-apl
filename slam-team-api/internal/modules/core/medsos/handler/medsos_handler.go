// Package handler exposes the medsos (social media links per anggota) HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/medsos/dto"
	"slam-team-api/internal/modules/core/medsos/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// MedsosHandler holds the Gin handlers for medsos CRUD.
type MedsosHandler struct {
	svc *service.MedsosService
}

// NewMedsosHandler creates a handler bound to the service.
func NewMedsosHandler(svc *service.MedsosService) *MedsosHandler {
	return &MedsosHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /medsos — paginated, filterable, sortable.
func (h *MedsosHandler) List(c *gin.Context) {
	var query dto.ListMedsosQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	cakupan := c.GetString("cakupan:medsos.read")

	items, total, err := h.svc.List(c.Request.Context(), query, actor, cakupan)
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

	response.OK(c, pagination.Paginated[dto.MedsosResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  perPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /medsos/:id — single medsos.
func (h *MedsosHandler) Detail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	actor := actorFromCtx(c)
	cakupan := c.GetString("cakupan:medsos.read")

	resp, err := h.svc.FindByID(c.Request.Context(), id, actor, cakupan)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Create handles POST /medsos — insert a new medsos row.
func (h *MedsosHandler) Create(c *gin.Context) {
	var req dto.CreateMedsosReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	cakupan := c.GetString("cakupan:medsos.create")

	resp, err := h.svc.Create(c.Request.Context(), req, actor, cakupan)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /medsos/:id — update an existing medsos row.
func (h *MedsosHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateMedsosReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	cakupan := c.GetString("cakupan:medsos.update")

	resp, err := h.svc.Update(c.Request.Context(), id, req, actor, cakupan)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /medsos/:id — soft-delete.
func (h *MedsosHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	actor := actorFromCtx(c)
	cakupan := c.GetString("cakupan:medsos.delete")

	if err := h.svc.SoftDelete(c.Request.Context(), id, actor, cakupan); err != nil {
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

// AnggotaTersedia handles GET /medsos/anggota-tersedia — dropdown data.
func (h *MedsosHandler) AnggotaTersedia(c *gin.Context) {
	items, err := h.svc.AnggotaTersedia(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, items)
}
