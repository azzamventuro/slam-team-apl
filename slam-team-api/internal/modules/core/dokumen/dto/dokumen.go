package dto

import "time"

// ─── Request ───

// CreateDokumenReq — POST /dokumen.
type CreateDokumenReq struct {
	FileUUID   string  `json:"file_uuid"  binding:"required"`
	ReffType   string  `json:"reff_type"  binding:"required,max=50"`
	ReffID     int64   `json:"reff_id"    binding:"required,gt=0"`
	Kode       string  `json:"kode"       binding:"omitempty,max=50"`
	Tipe       string  `json:"tipe"       binding:"omitempty,max=50"`
	Format     string  `json:"format"     binding:"omitempty,max=20"`
	Jenis      *int    `json:"jenis"      binding:"omitempty"`
	Keterangan string  `json:"keterangan" binding:"omitempty,max=255"`
}

// UpdateDokumenReq — PUT /dokumen/:id (all fields optional; file_uuid
// optional — if provided, the file reference is swapped).
type UpdateDokumenReq struct {
	FileUUID   *string `json:"file_uuid"  binding:"omitempty"`
	ReffType   *string `json:"reff_type"  binding:"omitempty,max=50"`
	ReffID     *int64  `json:"reff_id"    binding:"omitempty,gt=0"`
	Kode       *string `json:"kode"       binding:"omitempty,max=50"`
	Tipe       *string `json:"tipe"       binding:"omitempty,max=50"`
	Format     *string `json:"format"     binding:"omitempty,max=20"`
	Jenis      *int    `json:"jenis"      binding:"omitempty"`
	Keterangan *string `json:"keterangan" binding:"omitempty,max=255"`
}

// ListDokumenQuery — GET /dokumen.
type ListDokumenQuery struct {
	Page     int    `form:"page"      binding:"omitempty,min=1"`
	PerPage  int    `form:"per_page"  binding:"omitempty,min=1,max=100"`
	Q        string `form:"q"         binding:"omitempty"`
	Sort     string `form:"sort"      binding:"omitempty"`
	ReffType string `form:"reff_type" binding:"omitempty,max=50"`
	ReffID   *int64 `form:"reff_id"   binding:"omitempty"`
	Jenis    *int   `form:"jenis"     binding:"omitempty"`
	Tipe     string `form:"tipe"      binding:"omitempty,max=50"`
	Format   string `form:"format"    binding:"omitempty,max=20"`
}

func (q *ListDokumenQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}
}

// ─── Response ───

// FileRef is a lightweight representation of the attached file — only the
// fields a client needs to fetch the document via GET /files/{uuid}/{varian}.
// mst_file.id is intentionally omitted.
type FileRef struct {
	UUID       string `json:"uuid"`
	NamaAsli   string `json:"nama_asli"`
	Ekstensi   string `json:"ekstensi"`
	MimeType   string `json:"mime_type"`
	UkuranByte int64  `json:"ukuran_byte"`
	IsPublik   bool   `json:"is_publik"`
}

// DokumenResp — single document item in list or detail.
type DokumenResp struct {
	ID         int64     `json:"id"`
	Kode       *string   `json:"kode,omitempty"`
	Tipe       *string   `json:"tipe,omitempty"`
	Format     *string   `json:"format,omitempty"`
	ReffType   string    `json:"reff_type"`
	ReffID     int64     `json:"reff_id"`
	Jenis      *int      `json:"jenis,omitempty"`
	Keterangan *string   `json:"keterangan,omitempty"`
	File       *FileRef  `json:"file,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	CreatedBy  *int64    `json:"created_by,omitempty"`
	ModifiedAt *time.Time `json:"modified_at,omitempty"`
}
