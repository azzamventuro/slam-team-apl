// Package handler binds requests, calls the service, and writes the envelope.
// It never touches *gorm.DB and never picks a status code for a service error.
package handler

import (
	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/pengaturan/dto"
	"slam-team-api/internal/modules/core/pengaturan/service"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// PengaturanHandler serves the settings endpoints.
type PengaturanHandler struct {
	svc *service.PengaturanService
}

// NewPengaturanHandler wires the handler to its service.
func NewPengaturanHandler(svc *service.PengaturanService) *PengaturanHandler {
	return &PengaturanHandler{svc: svc}
}

// List handles GET /pengaturan — every setting, grouped by grup.
func (h *PengaturanHandler) List(c *gin.Context) {
	groups, err := h.svc.List(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, groups)
}

// ListByGrup handles GET /pengaturan/:grup — one tab's settings, 404 when the
// grup has no rows.
func (h *PengaturanHandler) ListByGrup(c *gin.Context) {
	items, err := h.svc.ListByGrup(c.Request.Context(), c.Param("grup"))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, items)
}

// Update handles PUT /pengaturan — a bulk change, applied atomically.
func (h *PengaturanHandler) Update(c *gin.Context) {
	var req dto.UpdatePengaturanReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	groups, err := h.svc.Update(c.Request.Context(), actorFrom(c), req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, groups)
}

// Public handles GET /public/pengaturan. It is mounted without JWTAuth, so it
// takes no actor and the service returns only the is_publik rows.
func (h *PengaturanHandler) Public(c *gin.Context) {
	settings, err := h.svc.Publik(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, settings)
}

// actorFrom builds the service Actor from the verified claims plus the request
// fingerprint the audit row records. Claims are present because the route runs
// behind JWTAuth; a nil UserID would mean a system action, which this endpoint
// never is.
func actorFrom(c *gin.Context) service.Actor {
	a := service.Actor{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if cl := middleware.Claims(c); cl != nil {
		id := int64(cl.UserID)
		a.UserID = &id
		a.IsSuper = cl.IsSuper
	}
	return a
}
