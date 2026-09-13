// Package dto defines the request/response shapes for the penugasan module
// (api-endpoints §7).
package dto

// AssignReq is the JSON body for POST /jadwal/:id/peserta.
//
// sesi_id null = every session of the schedule. wajib_absen defaults to true
// and is forced false by the service when the anggota has no user account.
type AssignReq struct {
	AnggotaID    int64   `json:"anggota_id"    binding:"required,gt=0"`
	SesiID       *int64  `json:"sesi_id"       binding:"omitempty,gt=0"`
	PeranPeserta string  `json:"peran_peserta" binding:"omitempty,oneof=peserta pelatih panitia pengawas"`
	WajibAbsen   *bool   `json:"wajib_absen"`
	Keterangan   *string `json:"keterangan"`
}

// BulkAssignReq is the JSON body for POST /jadwal/:id/peserta/bulk. Every
// anggota gets the same peran / wajib_absen; duplicates (already assigned to
// this schedule) are skipped, not rejected.
type BulkAssignReq struct {
	AnggotaIDs   []int64 `json:"anggota_ids"   binding:"required,min=1,max=500,dive,gt=0"`
	SesiID       *int64  `json:"sesi_id"       binding:"omitempty,gt=0"`
	PeranPeserta string  `json:"peran_peserta" binding:"omitempty,oneof=peserta pelatih panitia pengawas"`
	WajibAbsen   *bool   `json:"wajib_absen"`
}

// BulkAssignResp is the payload of the bulk endpoint.
type BulkAssignResp struct {
	Ditugaskan int `json:"ditugaskan"`
	Dilewati   int `json:"dilewati"`
}

// ResponReq is the JSON body for PATCH /peserta/:id/respon — the assignee
// accepts or declines. `izin` is not a response: it is set by the izin module.
type ResponReq struct {
	StatusTugas string  `json:"status_tugas" binding:"required,oneof=diterima ditolak"`
	Keterangan  *string `json:"keterangan"`
}

// ListPesertaQuery is the query string for GET /jadwal/:id/peserta.
type ListPesertaQuery struct {
	SesiID      *int64 `form:"sesi_id"      binding:"omitempty,gt=0"`
	StatusTugas string `form:"status_tugas" binding:"omitempty,oneof=ditugaskan diterima ditolak izin"`
	WajibAbsen  *bool  `form:"wajib_absen"`
	Q           string `form:"q"            binding:"omitempty,max=150"`
}
