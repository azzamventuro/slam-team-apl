// Package dto defines the request/response shapes for the notifikasi module.
package dto

// ListNotifikasiQuery is the query string for GET /notifikasi. The inbox is
// implicitly the caller's; there is no user filter by design.
type ListNotifikasiQuery struct {
	Page    int `form:"page,default=1"      binding:"min=1"`
	PerPage int `form:"per_page,default=20" binding:"min=1,max=100"`
	// IsDibaca filters read (true) / unread (false); omitted = both.
	IsDibaca *bool `form:"is_dibaca"`
	// Tipe filters one tipe value (jadwal_ditugaskan, …).
	Tipe string `form:"tipe" binding:"omitempty,max=50"`
	// Arsip includes archived rows when true.
	Arsip bool `form:"arsip"`
}

// JumlahResp is the payload of GET /notifikasi/jumlah-belum-dibaca.
type JumlahResp struct {
	Jumlah int64 `json:"jumlah"`
}

// BacaSemuaResp is the payload of PATCH /notifikasi/baca-semua.
type BacaSemuaResp struct {
	Ditandai int64 `json:"ditandai"`
}
