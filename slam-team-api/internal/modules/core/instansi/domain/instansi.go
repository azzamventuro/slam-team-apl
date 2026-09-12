// Package instansi holds the GORM entity for Master Data Instansi / Sekolah.
// Table: mst_instansi.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// Instansi maps to mst_instansi — the institutional entity (school / organization)
// that scopes anggota membership.
type Instansi struct {
	ID                  int64      `gorm:"column:id;primaryKey" json:"id"`
	Kode                string     `gorm:"column:kode"          json:"kode"`
	Nama                string     `gorm:"column:nama"          json:"nama"`
	NamaClub            string     `gorm:"column:nama_club"     json:"nama_club"`
	Alamat              string     `gorm:"column:alamat"        json:"alamat"`
	NoTelepon           string     `gorm:"column:no_telepon"    json:"no_telepon"`
	LogoUtamaFileID     *int64     `gorm:"column:logo_utama_file_id"     json:"logo_utama_file_id,omitempty"`
	LogoTambahanFileID  *int64     `gorm:"column:logo_tambahan_file_id"  json:"logo_tambahan_file_id,omitempty"`
	TanggalBergabung    *time.Time `gorm:"column:tanggal_bergabung;type:date" json:"tanggal_bergabung,omitempty"`
	Status              int        `gorm:"column:status;default:1" json:"status"`

	// File UUIDs resolved via LEFT JOIN in queries (not stored in DB).
	LogoUtamaUUID     *string `gorm:"column:logo_utama_uuid"     json:"logo_utama_uuid,omitempty"`
	LogoTambahanUUID  *string `gorm:"column:logo_tambahan_uuid"  json:"logo_tambahan_uuid,omitempty"`
	JumlahAnggota     int64   `gorm:"column:jumlah_anggota"      json:"jumlah_anggota"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Instansi) TableName() string { return "mst_instansi" }
