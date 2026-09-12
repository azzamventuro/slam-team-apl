// Package domain holds the GORM entity for Master Data Dokumen (polymorphic
// document attachments). Table: mst_dokumen.
package domain

import (
	filedomain "slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/shared/model"
)

// ValidReffTypes is the whitelist of allowed reff_type values for the
// polymorphic owner reference.
var ValidReffTypes = map[string]bool{
	"anggota":  true,
	"instansi": true,
	"inorga":   true,
	"unit":     true,
	"prestasi": true,
	"kegiatan": true,
	"artikel":  true,
}

// Dokumen maps to mst_dokumen — one row per document attachment that can
// point to any entity via the polymorphic pair (reff_type, reff_id).
type Dokumen struct {
	ID         int64             `gorm:"column:id;primaryKey"              json:"id"`
	Kode       *string           `gorm:"column:kode"                       json:"kode,omitempty"`
	FileID     *int64            `gorm:"column:file_id"                    json:"-"`
	Tipe       *string           `gorm:"column:tipe"                       json:"tipe,omitempty"`
	Format     *string           `gorm:"column:format"                     json:"format,omitempty"`
	ReffID     int64             `gorm:"column:reff_id;not null"           json:"reff_id"`
	ReffType   string            `gorm:"column:reff_type;not null"         json:"reff_type"`
	Jenis      *int              `gorm:"column:jenis"                      json:"jenis,omitempty"`
	Keterangan *string           `gorm:"column:keterangan"                 json:"keterangan,omitempty"`

	// File is eager-loaded via foreign key FileID → mst_file.id.
	// The response embeds file metadata (uuid, nama_asli, etc.) but never
	// exposes mst_file.id.
	File *filedomain.File `gorm:"foreignKey:FileID" json:"file,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Dokumen) TableName() string { return "mst_dokumen" }
