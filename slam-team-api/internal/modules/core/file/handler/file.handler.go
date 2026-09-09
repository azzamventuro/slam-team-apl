// Package handler binds requests, calls the service, and writes the envelope.
// It never touches *gorm.DB and never picks a status code for a service error.
package handler

import (
	"net/http"
	"strconv"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/modules/core/file/dto"
	"slam-team-api/internal/modules/core/file/service"
	"slam-team-api/internal/shared/response"
	"slam-team-api/pkg/validator"

	"github.com/gin-gonic/gin"
)

// FileHandler serves the file layer's endpoints.
type FileHandler struct {
	svc *service.FileService
}

// NewFileHandler wires the handler to its service.
func NewFileHandler(svc *service.FileService) *FileHandler {
	return &FileHandler{svc: svc}
}

// Upload handles POST /files — one multipart/form-data upload.
//
// The order here is the rancangan Bab 6.6 checklist, and it is load-bearing:
//
//   - cap the body with MaxBytesReader first, so an oversized upload is stopped
//     at the socket instead of after it has been buffered;
//   - ParseMultipartForm BEFORE FormFile — FormFile on an unparsed request is
//     the classic "http: no such file" that looks like a client bug;
//   - every I/O error downstream is checked, so "berhasil diunggah" can never
//     be reported for bytes that did not land.
//
// The client sends FormData WITHOUT a manual Content-Type: the browser has to
// generate the multipart boundary itself.
func (h *FileHandler) Upload(c *gin.Context) {
	maxBytes := h.svc.MaxBytes()
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

	if err := c.Request.ParseMultipartForm(maxBytes); err != nil {
		response.Unprocess(c, "unggahan tidak terbaca", map[string]string{
			"file": "permintaan bukan multipart/form-data yang valid, atau melebihi batas " +
				strconv.FormatInt(maxBytes>>20, 10) + " MB",
		})
		return
	}

	var req dto.UploadReq
	if err := c.ShouldBind(&req); err != nil {
		response.Unprocess(c, "validasi gagal", validator.Explain(err))
		return
	}

	fh, err := c.FormFile("file")
	if err != nil {
		response.Unprocess(c, "validasi gagal", map[string]string{"file": "berkas wajib diunggah"})
		return
	}

	out, err := h.svc.Upload(c.Request.Context(), actorFrom(c), req, fh)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Created(c, out)
}

// Detail handles GET /files/:uuid — the file's metadata and the variants that
// exist so far. It is how a client learns that status_proses has moved from
// menunggu to selesai without fetching the bytes.
func (h *FileHandler) Detail(c *gin.Context) {
	out, err := h.svc.Detail(c.Request.Context(), actorFrom(c), c.Param("uuid"))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, out)
}

// Serve handles GET /files/:uuid/:varian — the ONLY way to read a private file.
//
// Private bytes are never exposed through a static folder, which is what makes
// the guessable URL harmless: this handler decides, per file, whether the
// caller may see it (rancangan Bab 6.1).
//
// c.File delegates to http.ServeFile, so Range requests, Last-Modified and
// conditional GETs work for free. Content-Type is set first because ServeFile
// only sniffs when the header is absent, and the variant row knows better than
// a sniffer what was actually written.
func (h *FileHandler) Serve(c *gin.Context) {
	varian := domain.Varian(c.Param("varian"))
	out, err := h.svc.Serve(c.Request.Context(), actorFrom(c), c.Param("uuid"), varian)
	if err != nil {
		response.FromError(c, err)
		return
	}

	if out.MimeType != "" {
		c.Header("Content-Type", out.MimeType)
	}
	if out.IsPublik {
		// Variants are immutable for the life of a uuid, so a long cache is
		// safe and keeps club artwork off the API on every page view.
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		// Private evidence must not be written to a shared cache, nor left in
		// the browser's disk cache after a logout.
		c.Header("Cache-Control", "private, no-store")
	}
	c.Header("Content-Disposition", "inline; filename*=UTF-8''"+urlEncode(out.NamaAsli))
	// Tells a client that asked for `medium` that it is looking at `original`
	// because the derived variants are still being written.
	c.Header("X-File-Varian", string(out.Varian))

	c.File(out.AbsPath)
}

// Delete handles DELETE /files/:uuid — a soft delete. The ownership rule
// (Mod/User may only delete what they uploaded) lives in the service, which is
// the only layer that knows whose file this is.
func (h *FileHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), actorFrom(c), c.Param("uuid")); err != nil {
		response.FromError(c, err)
		return
	}
	response.OK(c, nil)
}

// actorFrom builds the service Actor from the verified claims plus the request
// fingerprint the audit row records.
//
// Claims may be absent: the serve route is mounted with optional auth so public
// files work without a token. A nil UserID then means "anonymous", and the
// service refuses every private file on that basis.
func actorFrom(c *gin.Context) service.Actor {
	a := service.Actor{
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	if cl := middleware.Claims(c); cl != nil {
		id := int64(cl.UserID)
		a.UserID = &id
		a.IsSuper = cl.IsSuper
		a.RoleLevel = cl.RoleLevel
	}
	return a
}

// urlEncode percent-encodes a filename for the RFC 5987 Content-Disposition
// parameter, so a name with spaces or non-ASCII characters survives the header.
func urlEncode(s string) string {
	const hex = "0123456789ABCDEF"
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= 'a' && ch <= 'z', ch >= 'A' && ch <= 'Z', ch >= '0' && ch <= '9',
			ch == '-', ch == '.', ch == '_', ch == '~':
			out = append(out, ch)
		default:
			out = append(out, '%', hex[ch>>4], hex[ch&0x0f])
		}
	}
	return string(out)
}
