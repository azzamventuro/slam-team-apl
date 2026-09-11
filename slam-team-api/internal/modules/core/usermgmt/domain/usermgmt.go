// Package domain holds the GORM entities for the user-management module.
// These map to the users, user_role, and mst_role tables. The auth module's
// User entity is a PLACEHOLDER — these are the real definitions.
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// User maps to the users table — a real user account linked to an anggota.
// Password is never serialized to JSON.
type User struct {
	ID              int64      `gorm:"column:id;primaryKey"                json:"id"`
	AnggotaID       int64      `gorm:"column:anggota_id;not null"         json:"anggota_id"`
	Username        string     `gorm:"column:username;not null"           json:"username"`
	Email           string     `gorm:"column:email;not null"              json:"email"`
	Password        string     `gorm:"column:password"                    json:"-"`
	RoleID          int64      `gorm:"column:role_id;not null"            json:"role_id"`
	Timezone        string     `gorm:"column:timezone;default:Asia/Jakarta" json:"timezone"`
	IsAktif         bool       `gorm:"column:is_aktif;default:true"       json:"is_aktif"`
	LoginTerakhir   *time.Time `gorm:"column:login_terakhir"              json:"login_terakhir,omitempty"`
	PasswordDiubah  *time.Time `gorm:"column:password_diubah"             json:"password_diubah,omitempty"`
	GagalLogin      int        `gorm:"column:gagal_login;default:0"       json:"gagal_login"`
	TerkunciSampai  *time.Time `gorm:"column:terkunci_sampai"             json:"terkunci_sampai,omitempty"`

	// Joined fields (resolved via LEFT JOIN, not stored in users table).
	AnggotaNamaLengkap *string `gorm:"column:anggota_nama_lengkap" json:"anggota_nama_lengkap,omitempty"`
	AnggotaNoInduk     *string `gorm:"column:anggota_no_induk"     json:"anggota_no_induk,omitempty"`
	AnggotaInstansiID  *int64  `gorm:"column:anggota_instansi_id"  json:"anggota_instansi_id,omitempty"`
	AnggotaFotoUUID    *string `gorm:"column:anggota_foto_uuid"    json:"anggota_foto_uuid,omitempty"`
	RoleNama           string  `gorm:"column:role_nama"            json:"role_nama,omitempty"`
	RoleLevel          int     `gorm:"column:role_level"           json:"role_level,omitempty"`
	RoleIsSuper        bool    `gorm:"column:role_is_super"        json:"role_is_super,omitempty"`

	model.Audit
}

// TableName pins the table name for GORM.
func (User) TableName() string { return "users" }

// UserRole maps to the user_role junction table.
// NOTE: user_role has NO modified_at, is_deleted, deleted_at, deleted_by columns
// so we cannot embed model.Audit — only created_at and created_by are present.
type UserRole struct {
	ID            int64      `gorm:"column:id;primaryKey"       json:"id"`
	UserID        int64      `gorm:"column:user_id;not null"    json:"user_id"`
	RoleID        int64      `gorm:"column:role_id;not null"    json:"role_id"`
	IsUtama       bool       `gorm:"column:is_utama;default:true" json:"is_utama"`
	BerlakuSampai *time.Time `gorm:"column:berlaku_sampai"      json:"berlaku_sampai,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at"          json:"created_at"`
	CreatedBy     *int64     `gorm:"column:created_by"          json:"created_by,omitempty"`
}

// TableName pins the table name for GORM.
func (UserRole) TableName() string { return "user_role" }

// Role maps to mst_role — read-only for this module (roles managed by hakakses).
type Role struct {
	ID       int64  `gorm:"column:id;primaryKey"       json:"id"`
	Kode     string `gorm:"column:kode"                json:"kode"`
	Nama     string `gorm:"column:nama"                json:"nama"`
	Level    int    `gorm:"column:level"               json:"level"`
	IsSuper  bool   `gorm:"column:is_super"            json:"is_super"`
	IsAktif  bool   `gorm:"column:is_aktif"            json:"is_aktif"`
	IsDeleted bool  `gorm:"column:is_deleted"          json:"-"`
}

// TableName pins the table name for GORM.
func (Role) TableName() string { return "mst_role" }

// AnggotaLite is a minimal projection of anggota for the available-anggota dropdown.
type AnggotaLite struct {
	ID           int64  `gorm:"column:id"             json:"id"`
	NamaLengkap  string `gorm:"column:nama_lengkap"   json:"nama_lengkap"`
	NoInduk      string `gorm:"column:no_induk"       json:"no_induk"`
	InstansiID   int64  `gorm:"column:instansi_id"    json:"instansi_id"`
	FotoUUID     string `gorm:"column:foto_uuid"      json:"foto_uuid,omitempty"`
}
