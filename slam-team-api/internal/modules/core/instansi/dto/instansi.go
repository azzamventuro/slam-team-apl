// Package dto defines the request/response shapes for the instansi module.
package dto

// CreateInstansiReq is the JSON body for POST /instansi.
type CreateInstansiReq struct {
	Kode                string  `json:"kode"                  binding:"required,max=50"`
	Nama                string  `json:"nama"                  binding:"required,max=150"`
	NamaClub            string  `json:"nama_club"             binding:"max=150"`
	Alamat              string  `json:"alamat"`
	NoTelepon           string  `json:"no_telepon"            binding:"max=30"`
	LogoUtamaFileID     *int64  `json:"logo_utama_file_id"    binding:"omitempty,gt=0"`
	LogoTambahanFileID  *int64  `json:"logo_tambahan_file_id" binding:"omitempty,gt=0"`
	TanggalBergabung    *string `json:"tanggal_bergabung"     binding:"omitempty"`
	Status              *int    `json:"status"                binding:"omitempty,oneof=0 1"`
}

// UpdateInstansiReq is the JSON body for PUT /instansi/:id — full replace, same shape.
type UpdateInstansiReq = CreateInstansiReq

// InstansiResp is the standard response shape for a single instansi.
type InstansiResp struct {
	ID                  int64   `json:"id"`
	Kode                string  `json:"kode"`
	Nama                string  `json:"nama"`
	NamaClub            string  `json:"nama_club"`
	Alamat              string  `json:"alamat"`
	NoTelepon           string  `json:"no_telepon"`
	LogoUtamaFileID     *int64  `json:"logo_utama_file_id"`
	LogoUtamaUUID       *string `json:"logo_utama_uuid"`
	LogoTambahanFileID  *int64  `json:"logo_tambahan_file_id"`
	LogoTambahanUUID    *string `json:"logo_tambahan_uuid"`
	TanggalBergabung    *string `json:"tanggal_bergabung"`
	Status              int     `json:"status"`
	JumlahAnggota       int64   `json:"jumlah_anggota"`
	CreatedAt           string  `json:"created_at"`
	ModifiedAt          *string `json:"modified_at"`
}

// ListInstansiQuery is the query-string shape for GET /instansi.
type ListInstansiQuery struct {
	Page    int    `form:"page,default=1"      binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"`
	Status  *int   `form:"status"              binding:"omitempty,oneof=0 1"`
}
