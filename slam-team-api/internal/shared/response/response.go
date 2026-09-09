// Package response provides the single JSON envelope used by every endpoint,
// so clients see one consistent response shape.
package response

import (
	"errors"
	"net/http"

	"slam-team-api/internal/shared/apperr"
	"slam-team-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Body is the standard envelope: {success, message, data?, errors?}.
type Body struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

// OK writes 200 with data.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Success: true, Message: "OK", Data: data})
}

// Created writes 201 with data.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, Body{Success: true, Message: "created", Data: data})
}

// BadRequest writes 400.
func BadRequest(c *gin.Context, msg string, errs any) {
	c.JSON(http.StatusBadRequest, Body{Success: false, Message: msg, Errors: errs})
}

// Unprocess writes 422 — typically validation failures.
func Unprocess(c *gin.Context, msg string, errs any) {
	c.JSON(http.StatusUnprocessableEntity, Body{Success: false, Message: msg, Errors: errs})
}

// Unauthorized writes 401.
func Unauthorized(c *gin.Context, msg string) {
	c.JSON(http.StatusUnauthorized, Body{Success: false, Message: msg})
}

// Internal logs the real error and writes a generic 500 (no leak to client).
func Internal(c *gin.Context, err error) {
	logger.Error("internal error", zap.Error(err))
	c.JSON(http.StatusInternalServerError, Body{Success: false, Message: "internal server error"})
}

// Forbidden writes 403 — permission (RBAC) denials.
func Forbidden(c *gin.Context, msg string) {
	c.JSON(http.StatusForbidden, Body{Success: false, Message: msg})
}

// NotFound writes 404.
func NotFound(c *gin.Context, msg string) {
	c.JSON(http.StatusNotFound, Body{Success: false, Message: msg})
}

// Conflict writes 409 — uniqueness/state conflicts.
func Conflict(c *gin.Context, msg string) {
	c.JSON(http.StatusConflict, Body{Success: false, Message: msg})
}

// FromError maps a service error to its HTTP status. This is the only place
// the mapping lives: services return apperr sentinels (optionally wrapped for
// context) and never pick status codes themselves. Anything unrecognised is an
// Internal — which is also the single point where 500s are logged.
func FromError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		NotFound(c, err.Error())
	case errors.Is(err, apperr.ErrForbidden):
		Forbidden(c, err.Error())
	case errors.Is(err, apperr.ErrConflict):
		Conflict(c, err.Error())
	case errors.Is(err, apperr.ErrValidation):
		Unprocess(c, err.Error(), nil)
	case errors.Is(err, apperr.ErrUnauthorized):
		Unauthorized(c, err.Error())
	default:
		Internal(c, err)
	}
}
