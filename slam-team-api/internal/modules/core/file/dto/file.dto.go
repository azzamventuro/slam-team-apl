// Package dto holds the file layer's request and response shapes.
package dto

import (
	"slam-team-api/internal/modules/core/file/domain"
)

// UploadReq is the non-file part of the multipart/form-data upload. It is bound
// with ShouldBind (form tags), never ShouldBindJSON — the body is a multipart
// stream whose boundary the browser generates.
//
// The binary part is named "file" and is read with c.FormFile, not through this
// struct.
type UploadReq struct {
	// Kategori decides the disk folder, the accepted content types and the
	// default is_publik. The oneof list mirrors domain.Kategori.
	Kategori string `form:"kategori" binding:"required,oneof=foto_profil foto_formal absensi banner logo dokumen flyer kta"`
	// ReffType is the owning table (anggota, absensi, …). `modul` is accepted
	// as the alias the upload component sends; both land here.
	ReffType string `form:"reff_type" binding:"omitempty,max=50"`
	Modul    string `form:"modul"     binding:"omitempty,max=50"`
	// ReffID is the owning row. Optional: a file is often uploaded before the
	// row that will point at it exists.
	ReffID *int64 `form:"reff_id" binding:"omitempty,min=1"`
	// IsPublik overrides the kategori default when sent; nil means "derive".
	IsPublik *bool `form:"is_publik"`
}

// Owner returns the effective reff_type, preferring the explicit field over the
// `modul` alias.
func (r UploadReq) Owner() string {
	if r.ReffType != "" {
		return r.ReffType
	}
	return r.Modul
}

// VarianResp is one rendered size on the wire. url is the authenticated serve
// route, which works for public and private files alike; url_publik is only set
// for public files that may also be fetched directly.
type VarianResp struct {
	Varian     string  `json:"varian"`
	URL        string  `json:"url"`
	URLPublik  *string `json:"url_publik,omitempty"`
	LebarPx    *int    `json:"lebar_px"`
	TinggiPx   *int    `json:"tinggi_px"`
	UkuranByte int64   `json:"ukuran_byte"`
	MimeType   string  `json:"mime_type"`
	Kualitas   *int    `json:"kualitas,omitempty"`
}

// UploadResp is what POST /files returns, and also the body of a file detail
// read. status_proses is "menunggu" while the medium/low variants are still
// being written, so a client that needs them can re-read or simply ask for
// "original", which always exists by the time this is returned.
type UploadResp struct {
	UUID         string       `json:"uuid"`
	NamaAsli     string       `json:"nama_asli"`
	NamaSlug     string       `json:"nama_slug"`
	Ekstensi     string       `json:"ekstensi"`
	MimeType     string       `json:"mime_type"`
	UkuranByte   int64        `json:"ukuran_byte"`
	Kategori     string       `json:"kategori"`
	ReffType     *string      `json:"reff_type,omitempty"`
	ReffID       *int64       `json:"reff_id,omitempty"`
	IsPublik     bool         `json:"is_publik"`
	StatusProses string       `json:"status_proses"`
	HashSHA256   string       `json:"hash_sha256"`
	LebarPx      *int         `json:"lebar_px"`
	TinggiPx     *int         `json:"tinggi_px"`
	Variants     []VarianResp `json:"variants"`
	// URL is the one an owning form should display: the medium variant for an
	// image, the original for anything else.
	URL string `json:"url"`
}

// ServeURL builds the authenticated serve path for one variant. It is the only
// place the route shape is written, so changing the mount point is one edit.
func ServeURL(uuid string, varian domain.Varian) string {
	return "/api/v1/files/" + uuid + "/" + string(varian)
}
