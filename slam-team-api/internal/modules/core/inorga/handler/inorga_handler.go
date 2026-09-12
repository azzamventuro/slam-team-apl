package handler

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/inorga/dto"
	"slam-team-api/internal/modules/core/inorga/service"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"
)

type InorgaHandler struct{ svc *service.InorgaService }

func NewInorgaHandler(svc *service.InorgaService) *InorgaHandler {
	return &InorgaHandler{svc: svc}
}

// ── List ──

func (h *InorgaHandler) List(c *gin.Context) {
	var q dto.ListInorgaQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	out, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, out)
}

// ── Detail ──

func (h *InorgaHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	out, err := h.svc.FindByID(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, out)
}

// ── Create ──

func (h *InorgaHandler) Create(c *gin.Context) {
	var req dto.CreateInorgaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	actor := actorFrom(c)
	out, err := h.svc.Create(c.Request.Context(), req, actor)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, out)
}

// ── Update ──

func (h *InorgaHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	var req dto.UpdateInorgaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}
	actor := actorFrom(c)
	out, err := h.svc.Update(c.Request.Context(), id, req, actor)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, out)
}

// ── Delete ──

func (h *InorgaHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "id tidak valid", nil)
		return
	}
	actor := actorFrom(c)
	if err := h.svc.SoftDelete(c.Request.Context(), id, actor); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, gin.H{"message": fmt.Sprintf("periode inorga ID %d berhasil dihapus", id)})
}

// actorFrom extracts the Actor from gin context (Claims set by JWT middleware).
func actorFrom(c *gin.Context) service.Actor {
	claims := middleware.Claims(c)
	actor := service.Actor{IPAddress: c.ClientIP()}
	if c.Request != nil {
		actor.UserAgent = c.Request.UserAgent()
	}
	if claims != nil {
		actor.UserID = int64(claims.UserID)
	}
	return actor
}
