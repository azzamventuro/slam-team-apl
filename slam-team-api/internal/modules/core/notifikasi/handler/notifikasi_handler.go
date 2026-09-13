// Package handler exposes the inbox endpoints. Every handler resolves the
// recipient from the JWT (middleware.Claims) and nothing else: there is no
// user parameter anywhere, so one user can never address another's rows.
package handler

import (
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/notifikasi/dto"
	"slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/internal/shared/pagination"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// NotifikasiHandler holds the Gin handlers for the notifikasi module.
type NotifikasiHandler struct {
	svc *service.NotifikasiService
}

// NewNotifikasiHandler creates a handler bound to the service.
func NewNotifikasiHandler(svc *service.NotifikasiService) *NotifikasiHandler {
	return &NotifikasiHandler{svc: svc}
}

// List handles GET /notifikasi — the caller's inbox, paginated.
func (h *NotifikasiHandler) List(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	var query dto.ListNotifikasiQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), userID, query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	lq := pagination.ListQuery{Page: query.Page, PerPage: query.PerPage}
	response.OK(c, pagination.New(items, lq, total))
}

// JumlahBelumDibaca handles GET /notifikasi/jumlah-belum-dibaca → {jumlah}.
func (h *NotifikasiHandler) JumlahBelumDibaca(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	n, err := h.svc.JumlahBelumDibaca(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, dto.JumlahResp{Jumlah: n})
}

// Baca handles PATCH /notifikasi/:id/baca — mark one read.
func (h *NotifikasiHandler) Baca(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	n, err := h.svc.TandaiDibaca(c.Request.Context(), userID, id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, n)
}

// BacaSemua handles PATCH /notifikasi/baca-semua — mark every unread row.
func (h *NotifikasiHandler) BacaSemua(c *gin.Context) {
	userID, ok := callerID(c)
	if !ok {
		return
	}
	n, err := h.svc.TandaiSemuaDibaca(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, dto.BacaSemuaResp{Ditandai: n})
}

// callerID reads the recipient from the JWT. JWTAuth already rejected
// anonymous requests, so a missing claim here is a wiring bug, answered
// with 401 rather than a panic.
func callerID(c *gin.Context) (int64, bool) {
	cl := middleware.Claims(c)
	if cl == nil || cl.UserID == 0 {
		response.Unauthorized(c, "tidak terautentikasi")
		return 0, false
	}
	return int64(cl.UserID), true
}
