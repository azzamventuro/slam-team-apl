// Package unit holds the GORM entity for Master Data Unit (airsoft gun registry).
// Table: unit.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// Unit maps to unit — one row per airsoft gun owned by an anggota.
type Unit struct {
	ID                int64      `gorm:"column:id;primaryKey"                json:"id"`
	AnggotaID         int64      `gorm:"column:anggota_id;not null"         json:"anggota_id"`
	Kode              string     `gorm:"column:kode"                        json:"kode"`
	Model             string     `gorm:"column:model"                       json:"model"`
	Panjang           *float64   `gorm:"column:panjang"                     json:"panjang,omitempty"`
	PanjangInbar      *float64   `gorm:"column:panjang_inbar"               json:"panjang_inbar,omitempty"`
	Lebar             *float64   `gorm:"column:lebar"                       json:"lebar,omitempty"`
	Berat             *float64   `gorm:"column:berat"                       json:"berat,omitempty"`
	BeratBB           *float64   `gorm:"column:berat_bb"                    json:"berat_bb,omitempty"`
	FPS               *float64   `gorm:"column:fps"                         json:"fps,omitempty"`
	DeskripsiWarna    string     `gorm:"column:deskripsi_warna"             json:"deskripsi_warna"`
	FotoSampulFileID  *int64     `gorm:"column:foto_sampul_file_id"         json:"foto_sampul_file_id,omitempty"`
	Disetujui         bool       `gorm:"column:disetujui;default:false"     json:"disetujui"`
	DisetujuiOleh     *int64     `gorm:"column:disetujui_oleh"              json:"disetujui_oleh,omitempty"`
	DisetujuiPada     *time.Time `gorm:"column:disetujui_pada"              json:"disetujui_pada,omitempty"`

	// Joined fields (resolved via LEFT JOIN in queries, not stored in unit table).
	AnggotaNama        *string `gorm:"column:anggota_nama"         json:"anggota_nama,omitempty"`
	AnggotaNoInduk     *string `gorm:"column:anggota_no_induk"     json:"anggota_no_induk,omitempty"`
	FotoSampulUUID     *string `gorm:"column:foto_sampul_uuid"     json:"foto_sampul_uuid,omitempty"`
	DisetujuiOlehNama  *string `gorm:"column:disetujui_oleh_nama"  json:"disetujui_oleh_nama,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Unit) TableName() string { return "unit" }
