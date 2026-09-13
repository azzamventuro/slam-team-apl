// Package dto defines the request/response shapes for the izin module
// (api-endpoints §9).
package dto

import "slam-team-api/internal/shared/pagination"

// CreateIzinReq is the JSON body for POST /izin. The requester is always the
// anggota behind the caller's account — there is no anggota_id field.
//
// waktu_pulang_diminta ("HH:MM" or "HH:MM:SS") is required when
// jenis = pulang_cepat and rejected for every other jenis; the rule lives in
// the service because it spans two fields.
type CreateIzinReq struct {
	SesiID             int64   `json:"sesi_id"              binding:"required,gt=0"`
	Jenis              string  `json:"jenis"                binding:"required,oneof=izin sakit dinas pulang_cepat"`
	Alasan             string  `json:"alasan"               binding:"required,max=2000"`
	WaktuPulangDiminta *string `json:"waktu_pulang_diminta" binding:"omitempty,max=8"`
	LampiranFileID     *int64  `json:"lampiran_file_id"     binding:"omitempty,gt=0"`
}

// TolakReq is the JSON body for PATCH /izin/:id/tolak.
type TolakReq struct {
	CatatanPeninjau string `json:"catatan_peninjau" binding:"required,max=2000"`
}

// ListIzinQuery is the query string for GET /izin. The owner filter is not a
// query parameter: it comes from the caller's cakupan.
type ListIzinQuery struct {
	pagination.ListQuery
	Status   string `form:"status"    binding:"omitempty,oneof=menunggu disetujui ditolak"`
	SesiID   *int64 `form:"sesi_id"   binding:"omitempty,gt=0"`
	JadwalID *int64 `form:"jadwal_id" binding:"omitempty,gt=0"`
	Jenis    string `form:"jenis"     binding:"omitempty,oneof=izin sakit dinas pulang_cepat"`
}

// IzinResp is the response shape for one request.
type IzinResp struct {
	ID                 int64         `json:"id"`
	SesiID             int64         `json:"sesi_id"`
	SesiTanggal        *string       `json:"sesi_tanggal,omitempty"`
	SesiStatus         *string       `json:"sesi_status,omitempty"`
	JadwalID           *int64        `json:"jadwal_id,omitempty"`
	JadwalNama         *string       `json:"jadwal_nama,omitempty"`
	JadwalKode         *string       `json:"jadwal_kode,omitempty"`
	AnggotaID          int64         `json:"anggota_id"`
	AnggotaNama        *string       `json:"anggota_nama,omitempty"`
	NoInduk            *string       `json:"no_induk,omitempty"`
	Jenis              string        `json:"jenis"`
	Alasan             string        `json:"alasan"`
	WaktuPulangDiminta *string       `json:"waktu_pulang_diminta,omitempty"`
	Lampiran           *LampiranInfo `json:"lampiran,omitempty"`
	Status             string        `json:"status"`
	DiprosesOleh       *int64        `json:"diproses_oleh,omitempty"`
	DiprosesOlehNama   *string       `json:"diproses_oleh_nama,omitempty"`
	DiprosesPada       *string       `json:"diproses_pada,omitempty"`
	CatatanPeninjau    *string       `json:"catatan_peninjau,omitempty"`
	CreatedAt          string        `json:"created_at"`
	ModifiedAt         *string       `json:"modified_at,omitempty"`
}

// LampiranInfo points the client at the private attachment: the file is only
// ever served through the authenticated /files route.
type LampiranInfo struct {
	FileID int64  `json:"file_id"`
	UUID   string `json:"uuid"`
	URL    string `json:"url"`
}
