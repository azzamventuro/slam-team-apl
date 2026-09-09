// Package apperr holds the sentinel errors that services return to signal a
// failure class. Handlers never map status codes themselves — they hand the
// error to response.FromError, which is the single place the mapping lives.
package apperr

import "errors"

// Sentinels returned by services (and repositories, for ErrNotFound). Wrap them
// with fmt.Errorf("...: %w", apperr.ErrX) to add context; FromError uses
// errors.Is, so wrapped values still map to the right status.
var (
	ErrNotFound     = errors.New("data tidak ditemukan")
	ErrForbidden    = errors.New("tidak diizinkan")
	ErrConflict     = errors.New("data bentrok")
	ErrValidation   = errors.New("validasi gagal")
	ErrUnauthorized = errors.New("tidak terautentikasi")
)
