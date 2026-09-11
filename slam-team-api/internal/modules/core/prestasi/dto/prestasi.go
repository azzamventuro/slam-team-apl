// Package dto defines the request/response shapes for the prestasi module.
package dto

// ── Request DTOs ──

// CreatePrestasiReq is the JSON body for POST /prestasi.
type CreatePrestasiReq struct {
	AnggotaID        int64   `json:"anggota_id"         binding:"required,gt=0"`
	Kode             *string `json:"kode"               binding:"omitempty,max=50"`
	Peringkat        *string `json:"peringkat"          binding:"omitempty,max=50"`
	Tingkat          *string `json:"tingkat"            binding:"omitempty,max=50"`
	JudulKompetisi   string  `json:"judul_kompetisi"    binding:"required,max=200"`
	FlyerFileID      *int64  `json:"flyer_file_id"      binding:"omitempty,gt=0"`
	TanggalKompetisi *string `json:"tanggal_kompetisi"  binding:"omitempty,datetime=2006-01-02"`
	AlamatKompetisi  *string `json:"alamat_kompetisi"   binding:"omitempty"`
	FotoSampulFileID *int64  `json:"foto_sampul_file_id" binding:"omitempty,gt=0"`
	Keterangan       *string `json:"keterangan"         binding:"omitempty"`
}

// UpdatePrestasiReq is the JSON body for PUT /prestasi/:id.
type UpdatePrestasiReq struct {
	Kode             *string `json:"kode"               binding:"omitempty,max=50"`
	Peringkat        *string `json:"peringkat"          binding:"omitempty,max=50"`
	Tingkat          *string `json:"tingkat"            binding:"omitempty,max=50"`
	JudulKompetisi   *string `json:"judul_kompetisi"    binding:"omitempty,max=200"`
	FlyerFileID      *int64  `json:"flyer_file_id"      binding:"omitempty,gt=0"`
	TanggalKompetisi *string `json:"tanggal_kompetisi"  binding:"omitempty,datetime=2006-01-02"`
	AlamatKompetisi  *string `json:"alamat_kompetisi"   binding:"omitempty"`
	FotoSampulFileID *int64  `json:"foto_sampul_file_id" binding:"omitempty,gt=0"`
	Keterangan       *string `json:"keterangan"         binding:"omitempty"`
}

// ListPrestasiQuery is the query-string shape for GET /prestasi.
type ListPrestasiQuery struct {
	Page        int     `form:"page,default=1"        binding:"min=1"`
	PerPage     int     `form:"per_page,default=20"   binding:"min=1,max=100"`
	Q           string  `form:"q"`
	Sort        string  `form:"sort"`
	AnggotaID   *int64  `form:"anggota_id"            binding:"omitempty,gt=0"`
	Tingkat     *string `form:"tingkat"`
	Peringkat   *string `form:"peringkat"`
	TanggalDari *string `form:"tanggal_dari"           binding:"omitempty,datetime=2006-01-02"`
	TanggalSampai *string `form:"tanggal_sampai"       binding:"omitempty,datetime=2006-01-02"`
}

// ── Response DTOs ──

// FileRef is a nested reference for file data.
type FileRef struct {
	FileID   int64   `json:"file_id"`
	UUID     string  `json:"uuid"`
	IsPublik bool    `json:"is_publik"`
}

// AnggotaRef is a compact anggota reference for prestasi responses.
type AnggotaRef struct {
	ID            int64   `json:"id"`
	NamaLengkap   string  `json:"nama_lengkap"`
	NamaPanggilan *string `json:"nama_panggilan,omitempty"`
	NoInduk       *string `json:"no_induk,omitempty"`
}

// PrestasiResp is the standard response shape for a single prestasi.
type PrestasiResp struct {
	ID               int64       `json:"id"`
	AnggotaID        int64       `json:"anggota_id"`
	Anggota          AnggotaRef  `json:"anggota"`
	Kode             *string     `json:"kode,omitempty"`
	Peringkat        *string     `json:"peringkat,omitempty"`
	Tingkat          *string     `json:"tingkat,omitempty"`
	JudulKompetisi   string      `json:"judul_kompetisi"`
	Flyer            *FileRef    `json:"flyer,omitempty"`
	TanggalKompetisi *string     `json:"tanggal_kompetisi,omitempty"`
	AlamatKompetisi  *string     `json:"alamat_kompetisi,omitempty"`
	FotoSampul       *FileRef    `json:"foto_sampul,omitempty"`
	Keterangan       *string     `json:"keterangan,omitempty"`
	CreatedAt        string      `json:"created_at"`
	ModifiedAt       *string     `json:"modified_at,omitempty"`
}

// PublicPrestasiItem is the safe-for-public projection of a prestasi record.
type PublicPrestasiItem struct {
	JudulKompetisi   string     `json:"judul_kompetisi"`
	Peringkat        *string    `json:"peringkat,omitempty"`
	Tingkat          *string    `json:"tingkat,omitempty"`
	TanggalKompetisi *string    `json:"tanggal_kompetisi,omitempty"`
	AlamatKompetisi  *string    `json:"alamat_kompetisi,omitempty"`
	Anggota          AnggotaRef `json:"anggota"`
	FlyerUUID        *string    `json:"flyer_uuid,omitempty"`
	FotoSampulUUID   *string    `json:"foto_sampul_uuid,omitempty"`
}

// AnggotaOpsi is the compact dropdown option for anggota.
type AnggotaOpsi struct {
	ID          int64  `json:"id"`
	NamaLengkap string `json:"nama_lengkap"`
}
