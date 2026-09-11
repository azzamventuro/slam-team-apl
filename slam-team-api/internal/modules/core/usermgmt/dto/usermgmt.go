// Package dto defines the request/response shapes for the user-management module.
package dto

// ── Request DTOs ──

// CreateUserReq is the JSON body for POST /{surface}.
type CreateUserReq struct {
	AnggotaID int64   `json:"anggota_id" binding:"required,gt=0"`
	Username  string  `json:"username"   binding:"required,max=50"`
	Email     string  `json:"email"      binding:"required,email,max=150"`
	Password  string  `json:"password"   binding:"required,min=8"`
	RoleID    int64   `json:"role_id"    binding:"required,gt=0"`
	Timezone  *string `json:"timezone"   binding:"omitempty,max=50"`
	IsAktif   *bool   `json:"is_aktif"`
}

// UpdateUserReq is the JSON body for PUT /{surface}/:id — password optional.
type UpdateUserReq struct {
	Username   *string `json:"username"    binding:"omitempty,max=50"`
	Email      *string `json:"email"       binding:"omitempty,email,max=150"`
	Password   *string `json:"password"    binding:"omitempty,min=8"`
	RoleID     *int64  `json:"role_id"     binding:"omitempty,gt=0"`
	Timezone   *string `json:"timezone"    binding:"omitempty,max=50"`
	IsAktif    *bool   `json:"is_aktif"`
	BukaKunci  *bool   `json:"buka_kunci"`
}

// ListUserQuery is the query-string shape for GET /{surface}.
type ListUserQuery struct {
	Page       int    `form:"page,default=1"       binding:"min=1"`
	PerPage    int    `form:"per_page,default=20"  binding:"min=1,max=100"`
	Q          string `form:"q"`
	Sort       string `form:"sort"`
	IsAktif    *bool  `form:"is_aktif"`
	InstansiID *int64 `form:"instansi_id"          binding:"omitempty,gt=0"`
}

// ── Response DTOs ──

// UserResp is the standard response shape for a single user (no password).
type UserResp struct {
	ID             int64       `json:"id"`
	Anggota        AnggotaInfo `json:"anggota"`
	Username       string      `json:"username"`
	Email          string      `json:"email"`
	Role           RoleInfo    `json:"role"`
	Timezone       string      `json:"timezone"`
	IsAktif        bool        `json:"is_aktif"`
	LoginTerakhir  *string     `json:"login_terakhir,omitempty"`
	TerkunciSampai *string     `json:"terkunci_sampai,omitempty"`
	CreatedAt      string      `json:"created_at"`
	ModifiedAt     *string     `json:"modified_at,omitempty"`
}

// AnggotaInfo is the nested anggota data inside UserResp.
type AnggotaInfo struct {
	ID           int64   `json:"id"`
	NamaLengkap  string  `json:"nama_lengkap"`
	NoInduk      *string `json:"no_induk,omitempty"`
	InstansiID   int64   `json:"instansi_id"`
	FotoURL      *string `json:"foto_url,omitempty"`
}

// RoleInfo is the nested role data inside UserResp.
type RoleInfo struct {
	ID      int64  `json:"id"`
	Nama    string `json:"nama"`
	Level   int    `json:"level"`
	IsSuper bool   `json:"is_super"`
}

// RoleOpsi is a compact role option for dropdowns.
type RoleOpsi struct {
	ID    int64  `json:"id"`
	Nama  string `json:"nama"`
	Level int    `json:"level"`
}

// AnggotaOpsi is a compact anggota option for dropdowns.
type AnggotaOpsi struct {
	ID          int64  `json:"id"`
	NamaLengkap string `json:"nama_lengkap"`
	NoInduk     string `json:"no_induk,omitempty"`
}
