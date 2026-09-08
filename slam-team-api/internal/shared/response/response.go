// Package response provides the single JSON envelope used by every endpoint,
// so clients see one consistent response shape.
package response

import (
	"net/http"

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
