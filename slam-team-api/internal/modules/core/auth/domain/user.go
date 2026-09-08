// Package domain holds the auth module's GORM entities.
package domain

import "time"

// User is an application account. Table: users.
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
