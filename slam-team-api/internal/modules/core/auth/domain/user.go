// Package domain holds the auth module's GORM entities.
package domain

import "time"

// User is a PLACEHOLDER entity from the original scaffold and does NOT match
// the real schema: slamteam_db.users has anggota_id, username, role_id,
// timezone and the is_deleted/deleted_at/deleted_by triplet, and no `name`
// column. It is replaced wholesale by the auth module (Fase 0). No other module
// may reference it — model new entities on internal/shared/model.Audit instead.
//
// Table: users.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"column:name" json:"name"`
	Email     string    `gorm:"column:email;uniqueIndex" json:"email"`
	Password  string    `gorm:"column:password" json:"-"` // bcrypt hash, never serialized
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName pins the table name regardless of GORM's pluralizer.
func (User) TableName() string { return "users" }
