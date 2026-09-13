// Package domain holds the GORM entity for Master Data Lokasi.
// Table: mst_lokasi.
package domain

import (
	"slam-team-api/internal/shared/model"
)

// Lokasi maps to mst_lokasi — a training/activity location with a geofence
// (point + radius) and an IANA timezone. jadwal.lokasi_id references it
// (restrict) and snapshots latitude/longitude/radius_meter/timezone at
// creation time.
type Lokasi struct {
	ID          int64   `gorm:"column:id;primaryKey"       json:"id"`
	Kode        string  `gorm:"column:kode"                json:"kode"`
	Nama        string  `gorm:"column:nama"                json:"nama"`
	JenisLokasi string  `gorm:"column:jenis_lokasi"         json:"jenis_lokasi"`
	Alamat      string  `gorm:"column:alamat"              json:"alamat"`
	Latitude    float64 `gorm:"column:latitude"            json:"latitude"`
	Longitude   float64 `gorm:"column:longitude"           json:"longitude"`
	RadiusMeter int     `gorm:"column:radius_meter"        json:"radius_meter"`
	Timezone    string  `gorm:"column:timezone"            json:"timezone"`
	FotoFileID  *int64  `gorm:"column:foto_file_id"        json:"foto_file_id,omitempty"`
	Keterangan  string  `gorm:"column:keterangan"          json:"keterangan"`
	IsAktif     bool    `gorm:"column:is_aktif;default:true" json:"is_aktif"`

	// FotoUUID is resolved via LEFT JOIN in queries (not stored in DB).
	FotoUUID *string `gorm:"column:foto_uuid" json:"foto_uuid,omitempty"`
	// JumlahJadwal is the count of active (non-deleted) jadwal referencing this
	// lokasi, resolved by a subquery on reads; never written back.
	JumlahJadwal int64 `gorm:"column:jumlah_jadwal;->" json:"jumlah_jadwal"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Lokasi) TableName() string { return "mst_lokasi" }
