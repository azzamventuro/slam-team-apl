// Package domain holds the auth module's GORM entities: the users table
// (accounts) and sesi_login (refresh-token sessions). Both map the schema in
// slamteam_db.dbml exactly; a user's display name is NOT a column here — it
// lives on the linked anggota row and is joined in at read time.
package domain

import (
	"encoding/json"
	"time"

	"slam-team-api/internal/shared/model"
)

// User maps to the users table. Password is the bcrypt hash and is never
// serialized. Every user is an anggota (anggota_id) but not every anggota has
// an account; the column is a pointer because the DB allows NULL for the
// bootstrap window before the anggota FK is satisfied.
type User struct {
	ID             int64      `gorm:"column:id;primaryKey"                 json:"id"`
	AnggotaID      *int64     `gorm:"column:anggota_id"                   json:"anggota_id,omitempty"`
	Username       string     `gorm:"column:username;not null"            json:"username"`
	Email          string     `gorm:"column:email;not null"               json:"email"`
	Password       string     `gorm:"column:password"                     json:"-"`
	RoleID         int64      `gorm:"column:role_id;not null"             json:"role_id"`
	Timezone       string     `gorm:"column:timezone;default:Asia/Jakarta" json:"timezone"`
	IsAktif        bool       `gorm:"column:is_aktif;default:true"        json:"is_aktif"`
	LoginTerakhir  *time.Time `gorm:"column:login_terakhir"               json:"login_terakhir,omitempty"`
	PasswordDiubah *time.Time `gorm:"column:password_diubah"              json:"password_diubah,omitempty"`
	GagalLogin     int        `gorm:"column:gagal_login;default:0"        json:"gagal_login"`
	TerkunciSampai *time.Time `gorm:"column:terkunci_sampai"              json:"terkunci_sampai,omitempty"`

	// Joined fields (LEFT JOIN anggota / mst_file / mst_role) — not columns of
	// users. Filled by the repository's profile queries only.
	AnggotaNamaLengkap *string `gorm:"column:anggota_nama_lengkap;->" json:"anggota_nama_lengkap,omitempty"`
	AnggotaNoInduk     *string `gorm:"column:anggota_no_induk;->"     json:"anggota_no_induk,omitempty"`
	AnggotaFotoUUID    *string `gorm:"column:anggota_foto_uuid;->"    json:"anggota_foto_uuid,omitempty"`
	RoleNama           string  `gorm:"column:role_nama;->"            json:"role_nama,omitempty"`
	RoleLevel          int     `gorm:"column:role_level;->"           json:"role_level,omitempty"`
	RoleIsSuper        bool    `gorm:"column:role_is_super;->"        json:"role_is_super,omitempty"`

	model.Audit
}

// TableName pins the table name regardless of GORM's pluralizer.
func (User) TableName() string { return "users" }

// Locked reports whether the account is inside a lockout window.
func (u *User) Locked(now time.Time) bool {
	return u.TerkunciSampai != nil && u.TerkunciSampai.After(now)
}

// SesiLogin maps to sesi_login: one row per login, holding the opaque refresh
// token. Rotation and logout set dicabut_pada; the table has no soft-delete
// columns. Rows cascade away with their user.
type SesiLogin struct {
	ID            int64           `gorm:"column:id;primaryKey"          json:"id"`
	UserID        int64           `gorm:"column:user_id;not null"      json:"user_id"`
	RefreshToken  string          `gorm:"column:refresh_token;not null" json:"-"`
	IPAddress     *string         `gorm:"column:ip_address;type:inet"  json:"ip_address,omitempty"`
	UserAgent     *string         `gorm:"column:user_agent"            json:"user_agent,omitempty"`
	InfoPerangkat json.RawMessage `gorm:"column:info_perangkat;type:jsonb" json:"info_perangkat,omitempty"`
	BerlakuSampai time.Time       `gorm:"column:berlaku_sampai;not null" json:"berlaku_sampai"`
	DicabutPada   *time.Time      `gorm:"column:dicabut_pada"          json:"dicabut_pada,omitempty"`
	CreatedAt     time.Time       `gorm:"column:created_at"            json:"created_at"`
}

// TableName pins the table; GORM's pluraliser would guess wrong.
func (SesiLogin) TableName() string { return "sesi_login" }

// Active reports whether the session can still be exchanged for a new access
// token: not revoked and not past its expiry.
func (s *SesiLogin) Active(now time.Time) bool {
	return s.DicabutPada == nil && s.BerlakuSampai.After(now)
}
