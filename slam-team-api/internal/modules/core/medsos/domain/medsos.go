// Package domain holds the GORM entity for Master Data Medsos (social media links per anggota).
// Table: mst_medsos.
package domain

import (
	"slam-team-api/internal/shared/model"
)

// Medsos maps to mst_medsos — a social media link belonging to an anggota.
type Medsos struct {
	ID            int64   `gorm:"column:id;primaryKey"           json:"id"`
	AnggotaID     int64   `gorm:"column:anggota_id;not null"    json:"anggota_id"`
	Kode          *string `gorm:"column:kode"                    json:"kode,omitempty"`
	Tipe          int     `gorm:"column:tipe;default:0"          json:"tipe"`
	Icon          *string `gorm:"column:icon"                    json:"icon,omitempty"`
	JenisMedsos   string  `gorm:"column:jenis_medsos;not null"  json:"jenis_medsos"`
	KontenMedsos  string  `gorm:"column:konten_medsos;not null" json:"konten_medsos"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Medsos) TableName() string { return "mst_medsos" }
