// Package handler exposes the profile_club HTTP endpoints.
package handler

import (
	"math"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/profileclub/dto"
	"slam-team-api/internal/modules/core/profileclub/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// ProfileClubHandler holds the Gin handlers for profile_club CRUD.
type ProfileClubHandler struct {
	svc *service.ProfileClubService
}

// NewProfileClubHandler creates a handler bound to the service.
func NewProfileClubHandler(svc *service.ProfileClubService) *ProfileClubHandler {
	return &ProfileClubHandler{svc: svc}
}

// ── Handlers ──

// List handles GET /profile-club — paginated (practical: 0/1 rows).
func (h *ProfileClubHandler) List(c *gin.Context) {
	var query dto.ListProfileClubQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	query.Normalize()

	perPage := query.PerPage
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	listQ := service.ListQueryFromDTO(query)

	items, total, err := h.svc.List(c.Request.Context(), listQ)
	if err != nil {
		respondError(c, err)
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	if lastPage < 1 {
		lastPage = 1
	}

	response.OK(c, pagination.Paginated[dto.ProfileClubResp]{
		Items:    items,
		Page:     query.Page,
		PerPage:  perPage,
		Total:    total,
		LastPage: lastPage,
	})
}

// Detail handles GET /profile-club/:id — single profile club.
func (h *ProfileClubHandler) Detail(c *gin.Context) {
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

// Create handles POST /profile-club — insert the club profile.
func (h *ProfileClubHandler) Create(c *gin.Context) {
	var req dto.UpsertProfileClubReq
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

// Update handles PUT /profile-club/:id — update the club profile.
func (h *ProfileClubHandler) Update(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}

	var req dto.UpsertProfileClubReq
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

// Delete handles DELETE /profile-club/:id — soft-delete.
func (h *ProfileClubHandler) Delete(c *gin.Context) {
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

// PublicProfil handles GET /public/profil — public endpoint, no auth.
func (h *ProfileClubHandler) PublicProfil(c *gin.Context) {
	resp, err := h.svc.GetPublicProfil(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	response.OK(c, resp)
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
