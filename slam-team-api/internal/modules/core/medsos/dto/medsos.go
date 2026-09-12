// Package dto defines the request/response shapes for the medsos module.
package dto

// CreateMedsosReq is the JSON body for POST /medsos.
type CreateMedsosReq struct {
	AnggotaID     int64  `json:"anggota_id"     binding:"required"`
	JenisMedsos   string `json:"jenis_medsos"   binding:"required,max=50"`
	KontenMedsos  string `json:"konten_medsos"  binding:"required,max=255"`
	Icon          string `json:"icon"            binding:"omitempty,max=50"`
	Kode          string `json:"kode"            binding:"omitempty,max=50"`
	Tipe          int    `json:"tipe"            binding:"omitempty,min=0"`
}

// UpdateMedsosReq is the JSON body for PUT /medsos/:id — anggota_id is immutable.
type UpdateMedsosReq struct {
	JenisMedsos   string `json:"jenis_medsos"   binding:"required,max=50"`
	KontenMedsos  string `json:"konten_medsos"  binding:"required,max=255"`
	Icon          string `json:"icon"            binding:"omitempty,max=50"`
	Kode          string `json:"kode"            binding:"omitempty,max=50"`
	Tipe          int    `json:"tipe"            binding:"omitempty,min=0"`
}

// MedsosResp is the standard response shape for a single medsos.
type MedsosResp struct {
	ID           int64   `json:"id"`
	AnggotaID    int64   `json:"anggota_id"`
	Kode         *string `json:"kode,omitempty"`
	Tipe         int     `json:"tipe"`
	Icon         *string `json:"icon,omitempty"`
	JenisMedsos  string  `json:"jenis_medsos"`
	KontenMedsos string  `json:"konten_medsos"`
	CreatedAt    string  `json:"created_at"`
	CreatedBy    *int64  `json:"created_by,omitempty"`
	ModifiedAt   *string `json:"modified_at,omitempty"`
	ModifiedBy   *int64  `json:"modified_by,omitempty"`
}

// ListMedsosQuery is the query-string shape for GET /medsos.
type ListMedsosQuery struct {
	Page        int    `form:"page,default=1"      binding:"min=1"`
	PerPage     int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q           string `form:"q"`
	Sort        string `form:"sort"`
	AnggotaID   *int64 `form:"anggota_id"`
	JenisMedsos string `form:"jenis_medsos"`
}

// AnggotaOpsi is the compact dropdown option for anggota.
type AnggotaOpsi struct {
	ID          int64  `json:"id"`
	NamaLengkap string `json:"nama_lengkap"`
}
