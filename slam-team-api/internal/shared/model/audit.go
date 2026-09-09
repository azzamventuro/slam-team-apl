// Package model holds the column conventions shared by every persistent
// entity: the audit/soft-delete columns and the scopes that honour them.
package model

import (
	"time"

	"gorm.io/gorm"
)

// Audit is embedded by every domain entity backed by a data table. The schema
// uses explicit soft-delete columns (is_deleted + deleted_at + deleted_by), NOT
// gorm.DeletedAt — the column set differs from GORM's magic, so deletes and
// reads go through SoftDelete/NotDeleted instead.
type Audit struct {
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	CreatedBy  *int64     `gorm:"column:created_by" json:"created_by,omitempty"`
	ModifiedAt *time.Time `gorm:"column:modified_at" json:"modified_at,omitempty"`
	ModifiedBy *int64     `gorm:"column:modified_by" json:"modified_by,omitempty"`
	IsDeleted  bool       `gorm:"column:is_deleted;default:false" json:"-"`
	DeletedAt  *time.Time `gorm:"column:deleted_at" json:"-"`
	DeletedBy  *int64     `gorm:"column:deleted_by" json:"-"`
}

// NotDeleted is the read scope every list/get query must apply:
//
//	db.Model(&Anggota{}).Scopes(model.NotDeleted).Find(&rows)
func NotDeleted(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = false")
}

// SoftDeleteFields is the update map for a delete: rows are never hard-deleted.
// actor may be nil for system/job actions.
//
//	db.Model(&Anggota{}).Where("id = ?", id).Updates(model.SoftDeleteFields(actor))
func SoftDeleteFields(actor *int64) map[string]any {
	return map[string]any{
		"is_deleted": true,
		"deleted_at": time.Now().UTC(),
		"deleted_by": actor,
	}
}
