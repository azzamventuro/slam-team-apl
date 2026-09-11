// Package domain holds the GORM entity for Master Data Prestasi (achievements).
// Table: prestasi.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// Prestasi maps to prestasi — one row per competition achievement for an anggota.
type Prestasi struct {
	ID                int64      `gorm:"column:id;primaryKey"                json:"id"`
	AnggotaID         int64      `gorm:"column:anggota_id;not null"         json:"anggota_id"`
	Kode              *string    `gorm:"column:kode"                        json:"kode,omitempty"`
	Peringkat         *string    `gorm:"column:peringkat"                   json:"peringkat,omitempty"`
	Tingkat           *string    `gorm:"column:tingkat"                     json:"tingkat,omitempty"`
	JudulKompetisi    string     `gorm:"column:judul_kompetisi"             json:"judul_kompetisi"`
	FlyerFileID       *int64     `gorm:"column:flyer_file_id"               json:"flyer_file_id,omitempty"`
	TanggalKompetisi  *time.Time `gorm:"column:tanggal_kompetisi"           json:"tanggal_kompetisi,omitempty"`
	AlamatKompetisi   *string    `gorm:"column:alamat_kompetisi"            json:"alamat_kompetisi,omitempty"`
	FotoSampulFileID  *int64     `gorm:"column:foto_sampul_file_id"         json:"foto_sampul_file_id,omitempty"`
	Keterangan        *string    `gorm:"column:keterangan"                  json:"keterangan,omitempty"`

	// Joined fields (resolved via LEFT JOIN, not stored in prestasi table).
	AnggotaNama       *string `gorm:"column:anggota_nama"          json:"anggota_nama,omitempty"`
	AnggotaNoInduk    *string `gorm:"column:anggota_no_induk"      json:"anggota_no_induk,omitempty"`
	AnggotaPanggilan  *string `gorm:"column:anggota_nama_panggilan" json:"anggota_nama_panggilan,omitempty"`
	FlyerFileIDRef    *int64  `gorm:"column:flyer_file_id_ref"     json:"flyer_file_id_ref,omitempty"`
	FlyerUUID         *string `gorm:"column:flyer_uuid"            json:"flyer_uuid,omitempty"`
	FlyerIsPublik     *bool   `gorm:"column:flyer_is_publik"       json:"flyer_is_publik,omitempty"`
	FotoSampulUUID    *string `gorm:"column:foto_sampul_uuid"      json:"foto_sampul_uuid,omitempty"`
	FotoSampulIsPublik *bool  `gorm:"column:foto_sampul_is_publik" json:"foto_sampul_is_publik,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Prestasi) TableName() string { return "prestasi" }
