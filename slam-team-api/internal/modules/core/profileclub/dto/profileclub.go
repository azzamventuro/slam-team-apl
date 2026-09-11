package dto

import "time"

// ─── Request ───

// UpsertProfileClubReq — POST / PUT /profile-club.
type UpsertProfileClubReq struct {
	Nama              string  `json:"nama"                binding:"required,max=150"`
	Singkatan         string  `json:"singkatan"           binding:"omitempty,max=50"`
	BannerFileID      *int64  `json:"banner_file_id"      binding:"omitempty"`
	LogoSimpleFileID  *int64  `json:"logo_simple_file_id"  binding:"omitempty"`
	LogoBesarFileID   *int64  `json:"logo_besar_file_id"   binding:"omitempty"`
	Alamat            string  `json:"alamat"              binding:"omitempty"`
	Keterangan        string  `json:"keterangan"          binding:"omitempty"`
}

// ListProfileClubQuery — GET /profile-club (practical: 0/1 rows).
type ListProfileClubQuery struct {
	Page    int    `form:"page"     binding:"omitempty,min=1"`
	PerPage int    `form:"per_page" binding:"omitempty,min=1,max=100"`
	Q       string `form:"q"        binding:"omitempty"`
	Sort    string `form:"sort"     binding:"omitempty"`
}

func (q *ListProfileClubQuery) Normalize() {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PerPage < 1 || q.PerPage > 100 {
		q.PerPage = 20
	}
}

// ─── Response ───

// ProfileClubResp — admin response for GET/POST/PUT.
type ProfileClubResp struct {
	ID                int64      `json:"id"`
	Nama              string     `json:"nama"`
	Singkatan         string     `json:"singkatan,omitempty"`
	BannerFileID      *int64     `json:"banner_file_id,omitempty"`
	LogoSimpleFileID  *int64     `json:"logo_simple_file_id,omitempty"`
	LogoBesarFileID   *int64     `json:"logo_besar_file_id,omitempty"`
	BannerUUID        *string    `json:"banner_uuid,omitempty"`
	LogoSimpleUUID    *string    `json:"logo_simple_uuid,omitempty"`
	LogoBesarUUID     *string    `json:"logo_besar_uuid,omitempty"`
	Alamat            string     `json:"alamat,omitempty"`
	Keterangan        string     `json:"keterangan,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	CreatedBy         *int64     `json:"created_by,omitempty"`
	ModifiedAt        *time.Time `json:"modified_at,omitempty"`
}

// PublicProfilResp — public response for GET /public/profil (no id, no audit).
type PublicProfilResp struct {
	Nama             string              `json:"nama"`
	Singkatan        string              `json:"singkatan,omitempty"`
	Alamat           string              `json:"alamat,omitempty"`
	Keterangan       string              `json:"keterangan,omitempty"`
	BannerURL        string              `json:"banner_url,omitempty"`
	LogoSimpleURL    string              `json:"logo_simple_url,omitempty"`
	LogoBesarURL     string              `json:"logo_besar_url,omitempty"`
	Prestasi         []PrestasiPublik    `json:"prestasi"`
}

// PrestasiPublik is the public-safe subset of prestasi (for profile_club public).
type PrestasiPublik struct {
	JudulKompetisi   string  `json:"judul_kompetisi"`
	Peringkat        string  `json:"peringkat,omitempty"`
	Tingkat          string  `json:"tingkat,omitempty"`
	TanggalKompetisi *string `json:"tanggal_kompetisi,omitempty"`
}
