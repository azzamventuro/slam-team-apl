// Package domain holds the GORM entity for Master Data Inorga (kepengurusan).
// Table: mst_inorga.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// Inorga maps to mst_inorga — one row per organisational period (kepengurusan).
type Inorga struct {
	ID             int64      `gorm:"column:id;primaryKey"                json:"id"`
	Kode           *string    `gorm:"column:kode"                         json:"kode,omitempty"`
	Nama           string     `gorm:"column:nama;not null"                json:"nama"`
	LogoFileID     *int64     `gorm:"column:logo_file_id"                 json:"logo_file_id,omitempty"`
	BannerFileID   *int64     `gorm:"column:banner_file_id"               json:"banner_file_id,omitempty"`
	TanggalMulai   *time.Time `gorm:"column:tanggal_mulai"               json:"tanggal_mulai,omitempty"`
	TanggalSelesai *time.Time `gorm:"column:tanggal_selesai"              json:"tanggal_selesai,omitempty"`
	FileSKFileID   *int64     `gorm:"column:file_sk_file_id"              json:"file_sk_file_id,omitempty"`
	Konten         *string    `gorm:"column:konten"                       json:"konten,omitempty"`

	// Joined fields (resolved via LEFT JOINs).
	LogoUUID       *string `gorm:"column:logo_uuid"           json:"logo_uuid,omitempty"`
	BannerUUID     *string `gorm:"column:banner_uuid"         json:"banner_uuid,omitempty"`
	FileSKUUID     *string `gorm:"column:file_sk_uuid"        json:"file_sk_uuid,omitempty"`

	// Virtual: computed from tanggal_mulai / tanggal_selesai.
	Aktif bool `gorm:"-" json:"aktif"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Inorga) TableName() string { return "mst_inorga" }
