// Package dto defines the request/response shapes for the unit module.
package dto

// ── Request DTOs ──

// CreateUnitReq is the JSON body for POST /unit.
type CreateUnitReq struct {
	AnggotaID        int64   `json:"anggota_id"        binding:"required,gt=0"`
	Kode             string  `json:"kode"              binding:"omitempty,max=50"`
	Model            string  `json:"model"             binding:"omitempty,max=100"`
	Panjang          *float64 `json:"panjang"          binding:"omitempty,gte=0"`
	PanjangInbar     *float64 `json:"panjang_inbar"    binding:"omitempty,gte=0"`
	Lebar            *float64 `json:"lebar"            binding:"omitempty,gte=0"`
	Berat            *float64 `json:"berat"            binding:"omitempty,gte=0"`
	BeratBB          *float64 `json:"berat_bb"         binding:"omitempty,gte=0"`
	FPS              *float64 `json:"fps"              binding:"omitempty,gte=0"`
	DeskripsiWarna   string  `json:"deskripsi_warna"  binding:"omitempty,max=150"`
	FotoSampulUUID   *string `json:"foto_sampul_uuid"  binding:"omitempty"`
}

// UpdateUnitReq is the JSON body for PUT /unit/:id — approval fields allowed.
type UpdateUnitReq struct {
	Kode             string  `json:"kode"              binding:"omitempty,max=50"`
	Model            string  `json:"model"             binding:"omitempty,max=100"`
	Panjang          *float64 `json:"panjang"          binding:"omitempty,gte=0"`
	PanjangInbar     *float64 `json:"panjang_inbar"    binding:"omitempty,gte=0"`
	Lebar            *float64 `json:"lebar"            binding:"omitempty,gte=0"`
	Berat            *float64 `json:"berat"            binding:"omitempty,gte=0"`
	BeratBB          *float64 `json:"berat_bb"         binding:"omitempty,gte=0"`
	FPS              *float64 `json:"fps"              binding:"omitempty,gte=0"`
	DeskripsiWarna   string  `json:"deskripsi_warna"  binding:"omitempty,max=150"`
	FotoSampulUUID   *string `json:"foto_sampul_uuid"  binding:"omitempty"`
	Disetujui        *bool   `json:"disetujui"         binding:"omitempty"`
}

// ListUnitQuery is the query-string shape for GET /unit.
type ListUnitQuery struct {
	Page     int    `form:"page,default=1"      binding:"min=1"`
	PerPage  int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q        string `form:"q"`
	AnggotaID *int64 `form:"anggota_id"         binding:"omitempty,gt=0"`
	Disetujui *bool  `form:"disetujui"          binding:"omitempty"`
	Sort     string `form:"sort"`
}

// ── Response DTOs ──

// UnitResp is the standard response shape for a single unit.
type UnitResp struct {
	ID               int64    `json:"id"`
	AnggotaID        int64    `json:"anggota_id"`
	AnggotaNama      *string  `json:"anggota_nama,omitempty"`
	AnggotaNoInduk   *string  `json:"anggota_no_induk,omitempty"`
	Kode             string   `json:"kode"`
	Model            string   `json:"model"`
	Panjang          *float64 `json:"panjang,omitempty"`
	PanjangInbar     *float64 `json:"panjang_inbar,omitempty"`
	Lebar            *float64 `json:"lebar,omitempty"`
	Berat            *float64 `json:"berat,omitempty"`
	BeratBB          *float64 `json:"berat_bb,omitempty"`
	FPS              *float64 `json:"fps,omitempty"`
	DeskripsiWarna   string   `json:"deskripsi_warna"`
	FotoSampul       *FotoSampulInfo `json:"foto_sampul,omitempty"`
	Disetujui        bool     `json:"disetujui"`
	DisetujuiOleh    *int64   `json:"disetujui_oleh,omitempty"`
	DisetujuiOlehNama *string `json:"disetujui_oleh_nama,omitempty"`
	DisetujuiPada    *string  `json:"disetujui_pada,omitempty"`
	CreatedAt        string   `json:"created_at"`
	ModifiedAt       *string  `json:"modified_at,omitempty"`
}

// FotoSampulInfo contains the foto sampul UUID and URL for the response.
type FotoSampulInfo struct {
	UUID string `json:"uuid"`
	URL  string `json:"url"`
}

// AnggotaOpsi is a compact anggota option for dropdowns.
type AnggotaOpsi struct {
	ID          int64  `json:"id"`
	NamaLengkap string `json:"nama_lengkap"`
	NoInduk     string `json:"no_induk"`
}
