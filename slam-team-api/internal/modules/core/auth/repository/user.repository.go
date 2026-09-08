// Package repository is the auth module's GORM data-access layer.
package repository

import (
	"slam-team-api/internal/modules/core/auth/domain"

	"gorm.io/gorm"
)

// UserRepository reads and writes users.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByEmail returns the user with the given email, or gorm.ErrRecordNotFound.
func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
