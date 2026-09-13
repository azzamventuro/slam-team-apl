// Package domain holds the GORM entity for excuse requests (absensi_izin):
// a member asking to be excused from one session, or to leave it early.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// JenisIzin mirrors the jenis_izin PG enum.
type JenisIzin string

const (
	JenisIzinBiasa   JenisIzin = "izin"
	JenisSakit       JenisIzin = "sakit"
	JenisDinas       JenisIzin = "dinas"
	JenisPulangCepat JenisIzin = "pulang_cepat"
)

// StatusIzin mirrors the status_izin PG enum.
type StatusIzin string

const (
	StatusMenunggu  StatusIzin = "menunggu"
	StatusDisetujui StatusIzin = "disetujui"
	StatusDitolak   StatusIzin = "ditolak"
)

// Notification tipe values produced by this module. The notifikasi.tipe
// column is a varchar, so the constants live next to their producer.
const (
	NotifTipeDisetujui = "izin_disetujui"
	NotifTipeDitolak   = "izin_ditolak"
)

// Izin maps to absensi_izin. It is keyed by anggota_id (the member behind
// the requester's account), tied to one jadwal_sesi; jadwal_id is copied
// from the session at create time.
//
// WaktuPulangDiminta is the PG `time` column kept as "HH:MM:SS" text (the
// same convention as jadwal.jam_mulai) and is only meaningful for
// jenis = pulang_cepat.
//
// The absensi linkage — an approved izin becomes the member's
// status_kehadiran on the session — is owned by the absensi module
// (absensi.izin_id → absensi_izin.id), not by this entity.
type Izin struct {
	ID                 int64      `gorm:"column:id;primaryKey"`
	SesiID             int64      `gorm:"column:sesi_id"`
	JadwalID           *int64     `gorm:"column:jadwal_id"`
	AnggotaID          int64      `gorm:"column:anggota_id"`
	Jenis              JenisIzin  `gorm:"column:jenis"`
	Alasan             string     `gorm:"column:alasan"`
	WaktuPulangDiminta *string    `gorm:"column:waktu_pulang_diminta;type:time"`
	LampiranFileID     *int64     `gorm:"column:lampiran_file_id"`
	Status             StatusIzin `gorm:"column:status;default:menunggu"`
	DiprosesOleh       *int64     `gorm:"column:diproses_oleh"`
	DiprosesPada       *time.Time `gorm:"column:diproses_pada"`
	CatatanPeninjau    *string    `gorm:"column:catatan_peninjau"`

	// Read-only projections resolved by the repository's read query, never
	// written back (GORM "->" permission).
	AnggotaNama      *string    `gorm:"column:anggota_nama;->"`
	NoInduk          *string    `gorm:"column:no_induk;->"`
	SesiTanggal      *time.Time `gorm:"column:sesi_tanggal;->;type:date"`
	SesiStatus       *string    `gorm:"column:sesi_status;->"`
	JadwalNama       *string    `gorm:"column:jadwal_nama;->"`
	JadwalKode       *string    `gorm:"column:jadwal_kode;->"`
	LampiranUUID     *string    `gorm:"column:lampiran_uuid;->"`
	DiprosesOlehNama *string    `gorm:"column:diproses_oleh_nama;->"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Izin) TableName() string { return "absensi_izin" }

// ProjectionColumns are the "->" fields, named explicitly on writes so GORM
// never tries to insert them.
var ProjectionColumns = []string{
	"anggota_nama", "no_induk", "sesi_tanggal", "sesi_status",
	"jadwal_nama", "jadwal_kode", "lampiran_uuid", "diproses_oleh_nama",
}

// MenjadiKehadiran reports whether an approved request of this jenis replaces
// the member's attendance for the whole session (izin / sakit / dinas). A
// pulang_cepat member still attends — they only leave early.
func (j JenisIzin) MenjadiKehadiran() bool {
	return j == JenisIzinBiasa || j == JenisSakit || j == JenisDinas
}
