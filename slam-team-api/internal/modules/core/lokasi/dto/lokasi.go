// Package dto defines the request/response shapes for the lokasi module.
package dto

// CreateLokasiReq is the JSON body for POST /lokasi.
type CreateLokasiReq struct {
	Kode        string  `json:"kode"         binding:"required,max=50"`
	Nama        string  `json:"nama"         binding:"required,max=150"`
	JenisLokasi string  `json:"jenis_lokasi" binding:"omitempty,max=50"`
	Alamat      string  `json:"alamat"`
	Latitude    float64 `json:"latitude"     binding:"required,latitude"`
	Longitude   float64 `json:"longitude"    binding:"required,longitude"`
	RadiusMeter int     `json:"radius_meter" binding:"required,min=1,max=100000"`
	Timezone    string  `json:"timezone"     binding:"required"`
	FotoFileID  *int64  `json:"foto_file_id" binding:"omitempty,gt=0"`
	Keterangan  string  `json:"keterangan"`
	IsAktif     *bool   `json:"is_aktif"     binding:"omitempty"`
}

// UpdateLokasiReq is the JSON body for PUT /lokasi/:id — full replace, same shape.
type UpdateLokasiReq = CreateLokasiReq

// LokasiResp is the standard response shape for a single lokasi.
type LokasiResp struct {
	ID          int64   `json:"id"`
	Kode        string  `json:"kode"`
	Nama        string  `json:"nama"`
	JenisLokasi string  `json:"jenis_lokasi"`
	Alamat      string  `json:"alamat"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	RadiusMeter int     `json:"radius_meter"`
	Timezone    string  `json:"timezone"`
	FotoFileID  *int64  `json:"foto_file_id"`
	FotoUUID    *string `json:"foto_uuid"`
	Keterangan  string  `json:"keterangan"`
	IsAktif     bool    `json:"is_aktif"`
	// JumlahJadwal is always 0 until 16-jadwal lands — see service.toResp.
	JumlahJadwal int64   `json:"jumlah_jadwal"`
	CreatedAt    string  `json:"created_at"`
	ModifiedAt   *string `json:"modified_at"`
}

// ListLokasiQuery is the query-string shape for GET /lokasi.
type ListLokasiQuery struct {
	Page    int    `form:"page,default=1"      binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"`
	IsAktif *bool  `form:"is_aktif" binding:"omitempty"`
}
