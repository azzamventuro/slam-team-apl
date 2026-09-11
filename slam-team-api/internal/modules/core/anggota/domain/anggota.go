// Package anggota holds the GORM entity for Master Data Anggota.
// Table: anggota.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// JenisAnggota is the enum type for anggota membership class.
type JenisAnggota string

const (
	JenisSiswa          JenisAnggota = "siswa"
	JenisDewasa         JenisAnggota = "dewasa"
	JenisSiswaKeDewasa  JenisAnggota = "siswa_ke_dewasa"
)

// StatusAnggota is the enum type for anggota status.
type StatusAnggota string

const (
	StatusAktif    StatusAnggota = "aktif"
	StatusNonAktif StatusAnggota = "non_aktif"
)

// Anggota maps to the anggota table — the core member record.
type Anggota struct {
	ID                int64          `gorm:"column:id;primaryKey"                 json:"id"`
	InstansiID        int64          `gorm:"column:instansi_id;not null"          json:"instansi_id"`
	WilayahID         *int64         `gorm:"column:wilayah_id"                   json:"wilayah_id,omitempty"`
	NoInduk           *string        `gorm:"column:no_induk"                     json:"no_induk,omitempty"`
	NamaLengkap       string         `gorm:"column:nama_lengkap;not null"        json:"nama_lengkap"`
	NamaPanggilan     *string        `gorm:"column:nama_panggilan"               json:"nama_panggilan,omitempty"`
	FotoProfilFileID  *int64         `gorm:"column:foto_profil_file_id"          json:"foto_profil_file_id,omitempty"`
	FotoFormalFileID  *int64         `gorm:"column:foto_formal_file_id"          json:"foto_formal_file_id,omitempty"`
	JenisAnggota      JenisAnggota   `gorm:"column:jenis_anggota;not null"       json:"jenis_anggota"`
	JenisKelamin      *int           `gorm:"column:jenis_kelamin"                json:"jenis_kelamin,omitempty"`
	JenisIdentitas    *int           `gorm:"column:jenis_identitas"              json:"jenis_identitas,omitempty"`
	NoIdentitas       *string        `gorm:"column:no_identitas"                 json:"no_identitas,omitempty"`
	FileIdentitasFileID *int64       `gorm:"column:file_identitas_file_id"       json:"file_identitas_file_id,omitempty"`
	Pekerjaan         *string        `gorm:"column:pekerjaan"                    json:"pekerjaan,omitempty"`
	Alamat            *string        `gorm:"column:alamat"                       json:"alamat,omitempty"`
	KodePos           *string        `gorm:"column:kode_pos"                     json:"kode_pos,omitempty"`
	TempatLahir       *string        `gorm:"column:tempat_lahir"                 json:"tempat_lahir,omitempty"`
	TanggalLahir      time.Time      `gorm:"column:tanggal_lahir;type:date;not null" json:"tanggal_lahir"`
	TanggalBergabung  *time.Time     `gorm:"column:tanggal_bergabung;type:date"  json:"tanggal_bergabung,omitempty"`
	StatusAnggota     StatusAnggota  `gorm:"column:status_anggota;default:aktif" json:"status_anggota"`

	// Joined fields (resolved via LEFT JOIN, not stored in anggota table).
	InstansiNama      *string  `gorm:"column:instansi_nama"     json:"instansi_nama,omitempty"`
	WilayahNama       *string  `gorm:"column:wilayah_nama"      json:"wilayah_nama,omitempty"`
	FotoProfilUUID    *string  `gorm:"column:foto_profil_uuid"  json:"foto_profil_uuid,omitempty"`
	FotoFormalUUID    *string  `gorm:"column:foto_formal_uuid"  json:"foto_formal_uuid,omitempty"`
	FileIdentitasUUID *string  `gorm:"column:file_identitas_uuid" json:"file_identitas_uuid,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Anggota) TableName() string { return "anggota" }
