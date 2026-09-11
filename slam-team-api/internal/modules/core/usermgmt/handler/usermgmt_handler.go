// Package handler exposes the user-management (admin / moderator / user surface)
// HTTP endpoints. The same handler is instantiated three times — once per
// surface — each sharing the same service but bound to a different band.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/usermgmt/dto"
	"slam-team-api/internal/modules/core/usermgmt/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// UsermgmtHandler holds the Gin handlers for user-management CRUD.
type UsermgmtHandler struct {
	svc  *service.UsermgmtService
	band service.Band
}

// NewUsermgmtHandler creates a handler bound to the service and band.
func NewUsermgmtHandler(svc *service.UsermgmtService, band service.Band) *UsermgmtHandler {
	return &UsermgmtHandler{svc: svc, band: band}
}

// ── Handlers ──

// List handles GET /{surface} — paginated, filterable, sortable.
func (h *UsermgmtHandler) List(c *gin.Context) {
	var query dto.ListUserQuery
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

	actor := actorFromCtx(c)
	items, total, err := h.svc.List(c.Request.Context(), query, h.band)
	if err != nil {
		response.Internal(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(query.PerPage)))
	if lastPage < 1 {
		lastPage = 1
	}
	_ = actor // actor used for anti-escalation in service layer

	response.OK(c, pagination.Paginated[dto.UserResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  query.PerPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /{surface}/:id — single user with joins.
func (h *UsermgmtHandler) Detail(c *gin.Context) {
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

// Create handles POST /{surface} — insert a new user + user_role row.
func (h *UsermgmtHandler) Create(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Create(c.Request.Context(), req, h.band, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Created(c, resp)
}

// Update handles PUT /{surface}/:id — partial update.
func (h *UsermgmtHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	actor := actorFromCtx(c)
	resp, err := h.svc.Update(c.Request.Context(), id, req, h.band, actor)
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
}

// Delete handles DELETE /{surface}/:id — soft-delete with guard checks.
func (h *UsermgmtHandler) Delete(c *gin.Context) {
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

// AvailableRoles handles GET /{surface}/peran-tersedia — roles the actor can
// assign within this surface's band.
func (h *UsermgmtHandler) AvailableRoles(c *gin.Context) {
	actor := actorFromCtx(c)
	roles, err := h.svc.AvailableRoles(c.Request.Context(), h.band, actor.RoleLevel)
	if err != nil {
		response.Internal(c, err)
		return
	}
	response.OK(c, roles)
}

// AvailableAnggota handles GET /{surface}/anggota-tersedia — anggota without
// user accounts, optionally filtered by ?q= and ?instansi_id=.
func (h *UsermgmtHandler) AvailableAnggota(c *gin.Context) {
	q := c.Query("q")
	var instansiID *int64
	if s := c.Query("instansi_id"); s != "" {
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			instansiID = &v
		}
	}

	items, err := h.svc.AvailableAnggota(c.Request.Context(), q, instansiID)
	if err != nil {
		response.Internal(c, err)
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
	var userID int64
	var roleLevel int
	var isSuper bool
	if cl != nil {
		userID = int64(cl.UserID)
		roleLevel = cl.RoleLevel
		isSuper = cl.IsSuper
	}
	return service.Actor{
		UserID:    userID,
		RoleLevel: roleLevel,
		IsSuper:   isSuper,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
}

func respondError(c *gin.Context, err error) {
	response.FromError(c, err)
}
