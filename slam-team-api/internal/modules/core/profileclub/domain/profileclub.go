// Package domain holds the GORM entity for the club's profile (singleton).
// Table: profile_club (NO mst_ prefix).
package domain

import (
	"slam-team-api/internal/shared/model"
)

// ProfileClub maps to profile_club — the single active row holding the
// club's branding identity (banner, logos, name, address, tagline).
type ProfileClub struct {
	ID                int64  `gorm:"column:id;primaryKey" json:"id"`
	Nama              string `gorm:"column:nama;not null"  json:"nama"`
	Singkatan         string `gorm:"column:singkatan"     json:"singkatan,omitempty"`
	BannerFileID      *int64 `gorm:"column:banner_file_id"       json:"banner_file_id,omitempty"`
	LogoSimpleFileID  *int64 `gorm:"column:logo_simple_file_id"  json:"logo_simple_file_id,omitempty"`
	LogoBesarFileID   *int64 `gorm:"column:logo_besar_file_id"   json:"logo_besar_file_id,omitempty"`
	Alamat            string `gorm:"column:alamat"         json:"alamat,omitempty"`
	Keterangan        string `gorm:"column:keterangan"     json:"keterangan,omitempty"`

	// File UUIDs resolved via LEFT JOINs (not stored in DB).
	BannerUUID      *string `gorm:"column:banner_uuid"       json:"banner_uuid,omitempty"`
	LogoSimpleUUID  *string `gorm:"column:logo_simple_uuid"  json:"logo_simple_uuid,omitempty"`
	LogoBesarUUID   *string `gorm:"column:logo_besar_uuid"   json:"logo_besar_uuid,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM — must be exactly "profile_club".
func (ProfileClub) TableName() string { return "profile_club" }
