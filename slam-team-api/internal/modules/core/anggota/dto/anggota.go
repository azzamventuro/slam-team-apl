// Package dto defines the request/response shapes for the anggota module.
package dto

// CreateAnggotaReq is the JSON body for POST /anggota.
type CreateAnggotaReq struct {
	InstansiID          int64   `json:"instansi_id"           binding:"required,gt=0"`
	WilayahID           *int64  `json:"wilayah_id"            binding:"omitempty,gt=0"`
	NamaLengkap         string  `json:"nama_lengkap"          binding:"required,max=150"`
	NamaPanggilan       *string `json:"nama_panggilan"        binding:"omitempty,max=50"`
	FotoProfilFileID    *int64  `json:"foto_profil_file_id"   binding:"omitempty,gt=0"`
	FotoFormalFileID    *int64  `json:"foto_formal_file_id"   binding:"omitempty,gt=0"`
	JenisAnggota        string  `json:"jenis_anggota"         binding:"required,oneof=siswa dewasa siswa_ke_dewasa"`
	JenisKelamin        *int    `json:"jenis_kelamin"         binding:"omitempty,oneof=1 2"`
	JenisIdentitas      *int    `json:"jenis_identitas"       binding:"omitempty"`
	NoIdentitas         *string `json:"no_identitas"          binding:"omitempty,max=50"`
	FileIdentitasFileID *int64  `json:"file_identitas_file_id" binding:"omitempty,gt=0"`
	Pekerjaan           *string `json:"pekerjaan"             binding:"omitempty,max=100"`
	Alamat              *string `json:"alamat"`
	KodePos             *string `json:"kode_pos"              binding:"omitempty,max=10"`
	TempatLahir         *string `json:"tempat_lahir"          binding:"omitempty,max=100"`
	TanggalLahir        string  `json:"tanggal_lahir"         binding:"required"`
	TanggalBergabung    *string `json:"tanggal_bergabung"     binding:"omitempty"`
	StatusAnggota       *string `json:"status_anggota"        binding:"omitempty,oneof=aktif non_aktif"`
}

// UpdateAnggotaReq is the JSON body for PUT /anggota/:id — identical shape.
type UpdateAnggotaReq = CreateAnggotaReq

// AnggotaResp is the standard response shape for a single anggota.
type AnggotaResp struct {
	ID                int64   `json:"id"`
	InstansiID        int64   `json:"instansi_id"`
	InstansiNama      *string `json:"instansi_nama,omitempty"`
	WilayahID         *int64  `json:"wilayah_id,omitempty"`
	WilayahNama       *string `json:"wilayah_nama,omitempty"`
	NoInduk           *string `json:"no_induk,omitempty"`
	NamaLengkap       string  `json:"nama_lengkap"`
	NamaPanggilan     *string `json:"nama_panggilan,omitempty"`
	FotoProfilFileID  *int64  `json:"foto_profil_file_id,omitempty"`
	FotoProfilUUID    *string `json:"foto_profil_uuid,omitempty"`
	FotoFormalFileID  *int64  `json:"foto_formal_file_id,omitempty"`
	FotoFormalUUID    *string `json:"foto_formal_uuid,omitempty"`
	JenisAnggota      string  `json:"jenis_anggota"`
	JenisKelamin      *int    `json:"jenis_kelamin,omitempty"`
	JenisIdentitas    *int    `json:"jenis_identitas,omitempty"`
	NoIdentitas       *string `json:"no_identitas,omitempty"`
	FileIdentitasFileID *int64  `json:"file_identitas_file_id,omitempty"`
	FileIdentitasUUID *string `json:"file_identitas_uuid,omitempty"`
	Pekerjaan         *string `json:"pekerjaan,omitempty"`
	Alamat            *string `json:"alamat,omitempty"`
	KodePos           *string `json:"kode_pos,omitempty"`
	TempatLahir       *string `json:"tempat_lahir,omitempty"`
	TanggalLahir      string  `json:"tanggal_lahir"`
	TanggalBergabung  *string `json:"tanggal_bergabung,omitempty"`
	StatusAnggota     string  `json:"status_anggota"`
	CreatedAt         string  `json:"created_at"`
	ModifiedAt        *string `json:"modified_at,omitempty"`
}

// ListAnggotaQuery is the query-string shape for GET /anggota.
type ListAnggotaQuery struct {
	Page       int     `form:"page,default=1"       binding:"min=1"`
	PerPage    int     `form:"per_page,default=20"  binding:"min=1,max=100"`
	Q          string  `form:"q"`
	InstansiID *int64  `form:"instansi_id"          binding:"omitempty,gt=0"`
	Jenis      *string `form:"jenis"                binding:"omitempty,oneof=siswa dewasa siswa_ke_dewasa"`
	Status     *string `form:"status"               binding:"omitempty,oneof=aktif non_aktif"`
	Sort       string  `form:"sort"`
}
