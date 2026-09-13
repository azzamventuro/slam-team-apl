// Package domain holds the GORM entity for the in-app inbox (notifikasi).
package domain

import "time"

// Prioritas mirrors the prioritas_notifikasi PG enum.
type Prioritas string

const (
	PrioritasRendah Prioritas = "rendah"
	PrioritasNormal Prioritas = "normal"
	PrioritasTinggi Prioritas = "tinggi"
)

// Tipe values written so far. The column is a varchar, not an enum, so later
// modules add their own constants next to their producer (izin_disetujui,
// absensi_dioverride, kta_diterbitkan, …) without a migration.
const (
	TipeJadwalDitugaskan = "jadwal_ditugaskan"
	TipeJadwalPengingat  = "jadwal_pengingat"
	TipeSesiDibatalkan   = "sesi_dibatalkan"
	TipeSistem           = "sistem"
)

// Warna values the client maps to a badge colour.
const (
	WarnaInfo       = "info"
	WarnaSukses     = "sukses"
	WarnaPeringatan = "peringatan"
	WarnaBahaya     = "bahaya"
)

// Notifikasi maps to the notifikasi table — one row per recipient (fan-out).
// It carries no soft-delete columns: rows are read, archived, or swept by the
// kedaluwarsa job; a user's inbox disappears with the user (FK CASCADE).
//
// UserID is never serialised: every read is already scoped to the caller, so
// echoing the recipient would only ever repeat the caller's own id.
type Notifikasi struct {
	ID              int64      `gorm:"column:id;primaryKey"                  json:"id"`
	UUID            string     `gorm:"column:uuid;default:gen_random_uuid()" json:"uuid"`
	UserID          int64      `gorm:"column:user_id"                        json:"-"`
	Tipe            string     `gorm:"column:tipe"                           json:"tipe"`
	Judul           string     `gorm:"column:judul"                          json:"judul"`
	Isi             *string    `gorm:"column:isi"                            json:"isi,omitempty"`
	Ikon            *string    `gorm:"column:ikon"                           json:"ikon,omitempty"`
	Warna           *string    `gorm:"column:warna"                          json:"warna,omitempty"`
	Route           *string    `gorm:"column:route"                          json:"route,omitempty"`
	ReffType        *string    `gorm:"column:reff_type"                      json:"reff_type,omitempty"`
	ReffID          *int64     `gorm:"column:reff_id"                        json:"reff_id,omitempty"`
	Prioritas       Prioritas  `gorm:"column:prioritas;default:normal"       json:"prioritas"`
	IsDibaca        bool       `gorm:"column:is_dibaca"                      json:"is_dibaca"`
	DibacaPada      *time.Time `gorm:"column:dibaca_pada"                    json:"dibaca_pada,omitempty"`
	IsDiarsipkan    bool       `gorm:"column:is_diarsipkan"                  json:"is_diarsipkan"`
	KedaluwarsaPada *time.Time `gorm:"column:kedaluwarsa_pada"               json:"kedaluwarsa_pada,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at;default:now()"       json:"created_at"`
	CreatedBy       *int64     `gorm:"column:created_by"                     json:"created_by,omitempty"`
}

// TableName pins the table name for GORM.
func (Notifikasi) TableName() string { return "notifikasi" }
