// Package handler exposes the unit (Master Data Unit / airsoft gun registry)
// HTTP endpoints. The handler resolves scope from middleware context and
// delegates all business logic to the service layer.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/unit/dto"
	"slam-team-api/internal/modules/core/unit/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// UnitHandler holds the Gin handlers for unit CRUD.
type UnitHandler struct {
	svc *service.UnitService
}

// NewUnitHandler creates a handler bound to the service.
func NewUnitHandler(svc *service.UnitService) *UnitHandler {
	return &UnitHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /unit — paginated, filterable, sortable.
func (h *UnitHandler) List(c *gin.Context) {
	var query dto.ListUnitQuery
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

	scope := middleware.GetCakupan(c, "unit.read")
	actor := actorFromCtx(c, h.svc)
	items, total, err := h.svc.List(c.Request.Context(), query, scope, actor)
	if err != nil {
		respondError(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(query.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.UnitResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /unit/:id — single unit with joins.
func (h *UnitHandler) Detail(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	scope := middleware.GetCakupan(c, "unit.read")
	actor := actorFromCtx(c, h.svc)
	resp, err := h.svc.FindByID(c.Request.Context(), id, scope, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Create handles POST /unit — insert a new unit row.
func (h *UnitHandler) Create(c *gin.Context) {
	var req dto.CreateUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	scope := middleware.GetCakupan(c, "unit.create")
	actor := actorFromCtx(c, h.svc)
	resp, err := h.svc.Create(c.Request.Context(), req, scope, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /unit/:id — full-replace update with approval toggle.
func (h *UnitHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateUnitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	scope := middleware.GetCakupan(c, "unit.update")
	actor := actorFromCtx(c, h.svc)
	resp, err := h.svc.Update(c.Request.Context(), id, req, scope, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /unit/:id — soft-delete.
func (h *UnitHandler) Delete(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	scope := middleware.GetCakupan(c, "unit.delete")
	actor := actorFromCtx(c, h.svc)
	if err := h.svc.SoftDelete(c.Request.Context(), id, scope, actor); err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, nil)
}

// AnggotaTersedia handles GET /unit/anggota-tersedia — dropdown data.
func (h *UnitHandler) AnggotaTersedia(c *gin.Context) {
	items, err := h.svc.AnggotaTersedia(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, items)
}

// ── Helpers ──

func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

func actorFromCtx(c *gin.Context, svc *service.UnitService) service.Actor {
	cl := middleware.Claims(c)
	actor := service.Actor{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if cl != nil {
		actor.UserID = int64(cl.UserID)
		// Look up anggota_id from users table (JWT claims don't carry it).
		anggotaID, _ := svc.LookupAnggotaID(c.Request.Context(), int64(cl.UserID))
		actor.AnggotaID = anggotaID
	}
	return actor
}

func respondError(c *gin.Context, err error) {
	response.FromError(c, err)
}
