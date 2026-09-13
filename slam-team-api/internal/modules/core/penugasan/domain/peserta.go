// Package domain holds the GORM entity for participant assignment
// (jadwal_peserta): which anggota is expected at which schedule/session.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// StatusTugas mirrors the status_tugas PG enum.
type StatusTugas string

const (
	TugasDitugaskan StatusTugas = "ditugaskan"
	TugasDiterima   StatusTugas = "diterima"
	TugasDitolak    StatusTugas = "ditolak"
	TugasIzin       StatusTugas = "izin"
)

// Peran values accepted for peran_peserta (a varchar, listed in the DBML).
const (
	PeranPeserta  = "peserta"
	PeranPelatih  = "pelatih"
	PeranPanitia  = "panitia"
	PeranPengawas = "pengawas"
)

// JadwalPeserta maps to jadwal_peserta. It is keyed by anggota_id, NOT
// user_id: members without an account can still be assigned. For them
// WajibAbsen is forced false at assignment time so the session-closing job
// (which only marks alfa where wajib_absen = true) never penalises someone
// who could not have checked in.
//
// SesiID nil = the assignment covers every session of the schedule.
type JadwalPeserta struct {
	ID             int64       `gorm:"column:id;primaryKey"           json:"id"`
	JadwalID       int64       `gorm:"column:jadwal_id"               json:"jadwal_id"`
	SesiID         *int64      `gorm:"column:sesi_id"                 json:"sesi_id"`
	AnggotaID      int64       `gorm:"column:anggota_id"              json:"anggota_id"`
	PeranPeserta   *string     `gorm:"column:peran_peserta"           json:"peran_peserta,omitempty"`
	WajibAbsen     bool        `gorm:"column:wajib_absen"             json:"wajib_absen"`
	StatusTugas    StatusTugas `gorm:"column:status_tugas;default:ditugaskan" json:"status_tugas"`
	DitugaskanOleh int64       `gorm:"column:ditugaskan_oleh"         json:"ditugaskan_oleh"`
	DitugaskanPada time.Time   `gorm:"column:ditugaskan_pada;default:now()" json:"ditugaskan_pada"`
	DiresponPada   *time.Time  `gorm:"column:direspon_pada"           json:"direspon_pada,omitempty"`
	Keterangan     *string     `gorm:"column:keterangan"              json:"keterangan,omitempty"`

	// Read-only projections resolved by the repository's list/detail query,
	// never written back (GORM "->" permission).
	AnggotaNama    *string    `gorm:"column:anggota_nama;->"    json:"anggota_nama,omitempty"`
	NoInduk        *string    `gorm:"column:no_induk;->"        json:"no_induk,omitempty"`
	AnggotaBerakun bool       `gorm:"column:anggota_berakun;->" json:"anggota_berakun"`
	SesiTanggal    *time.Time `gorm:"column:sesi_tanggal;->;type:date" json:"sesi_tanggal,omitempty"`

	model.SoftDelete
}

// TableName pins the table name for GORM.
func (JadwalPeserta) TableName() string { return "jadwal_peserta" }
