package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

func TestFromErrorMapsSentinels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		err  error
		want int
	}{
		{"not found", apperr.ErrNotFound, http.StatusNotFound},
		{"forbidden", apperr.ErrForbidden, http.StatusForbidden},
		{"conflict", apperr.ErrConflict, http.StatusConflict},
		{"validation", apperr.ErrValidation, http.StatusUnprocessableEntity},
		{"unauthorized", apperr.ErrUnauthorized, http.StatusUnauthorized},
		{"wrapped sentinel", fmt.Errorf("anggota %d: %w", 7, apperr.ErrNotFound), http.StatusNotFound},
		{"unknown error", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

			response.FromError(c, tc.err)

			if rec.Code != tc.want {
				t.Fatalf("status = %d, want %d", rec.Code, tc.want)
			}

			var body response.Body
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("body is not the standard envelope: %v", err)
			}
			if body.Success {
				t.Error("success = true on an error response")
			}
			if body.Message == "" {
				t.Error("message is empty")
			}
		})
	}
}

// A 500 must never carry the real error text — that leaks schema internals.
func TestInternalDoesNotLeak(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	response.FromError(c, errors.New(`pq: column "rahasia" does not exist`))

	if got := rec.Body.String(); got == "" || strings.Contains(got, "rahasia") {
		t.Errorf("500 body leaked the underlying error: %s", got)
	}
}
