package dto

import "time"

// ─── Request ───

// CreateInorgaReq — POST /inorga.
type CreateInorgaReq struct {
	Kode           string  `json:"kode"                          binding:"omitempty,max=50"`
	Nama           string  `json:"nama"                          binding:"required,max=150"`
	TanggalMulai   string  `json:"tanggal_mulai"                 binding:"required"`
	TanggalSelesai *string `json:"tanggal_selesai"               binding:"omitempty"`
	LogoFileID     *int64  `json:"logo_file_id"                  binding:"omitempty,gt=0"`
	BannerFileID   *int64  `json:"banner_file_id"                binding:"omitempty,gt=0"`
	FileSKFileID   *int64  `json:"file_sk_file_id"               binding:"omitempty,gt=0"`
	Konten         *string `json:"konten"                        binding:"omitempty"`
}

// UpdateInorgaReq — PUT /inorga/:id (all fields optional).
type UpdateInorgaReq struct {
	Kode           *string `json:"kode"                          binding:"omitempty,max=50"`
	Nama           *string `json:"nama"                          binding:"omitempty,max=150"`
	TanggalMulai   *string `json:"tanggal_mulai"                 binding:"omitempty"`
	TanggalSelesai *string `json:"tanggal_selesai"               binding:"omitempty"`
	LogoFileID     *int64  `json:"logo_file_id"                  binding:"omitempty"`
	BannerFileID   *int64  `json:"banner_file_id"                binding:"omitempty"`
	FileSKFileID   *int64  `json:"file_sk_file_id"               binding:"omitempty"`
	Konten         *string `json:"konten"                        binding:"omitempty"`
}

// ListInorgaQuery — GET /inorga.
type ListInorgaQuery struct {
	Page    int    `form:"page"      binding:"omitempty,min=1"`
	PerPage int    `form:"per_page"  binding:"omitempty,min=1,max=100"`
	Q       string `form:"q"         binding:"omitempty"`
	Sort    string `form:"sort"      binding:"omitempty"`
	Aktif   string `form:"aktif"     binding:"omitempty,oneof=true false"`
}

func (q *ListInorgaQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}
}

// ─── Response ───

// FileRef is a lightweight reference to a file (UUID + optional nama_asli).
type FileRef struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	NamaAsli  string `json:"nama_asli,omitempty"`
}

// InorgaResp — single inorga detail.
type InorgaResp struct {
	ID             int64      `json:"id"`
	Kode           *string    `json:"kode,omitempty"`
	Nama           string     `json:"nama"`
	TanggalMulai   *time.Time `json:"tanggal_mulai,omitempty"`
	TanggalSelesai *time.Time `json:"tanggal_selesai,omitempty"`
	Konten         *string    `json:"konten,omitempty"`
	Aktif          bool       `json:"aktif"`
	Logo           *FileRef   `json:"logo,omitempty"`
	Banner         *FileRef   `json:"banner,omitempty"`
	FileSK         *FileRef   `json:"file_sk,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ModifiedAt     *time.Time `json:"modified_at,omitempty"`
}

// InorgaListItem — lighter response for list (no konten).
type InorgaListItem struct {
	ID             int64      `json:"id"`
	Kode           *string    `json:"kode,omitempty"`
	Nama           string     `json:"nama"`
	TanggalMulai   *time.Time `json:"tanggal_mulai,omitempty"`
	TanggalSelesai *time.Time `json:"tanggal_selesai,omitempty"`
	Aktif          bool       `json:"aktif"`
	Logo           *FileRef   `json:"logo,omitempty"`
	Banner         *FileRef   `json:"banner,omitempty"`
	FileSK         *FileRef   `json:"file_sk,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
