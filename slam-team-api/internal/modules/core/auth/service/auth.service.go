// Package service holds the auth module's business logic.
package service

import (
	"errors"

	"slam-team-api/internal/modules/core/auth/domain"
	"slam-team-api/internal/modules/core/auth/dto"
	"slam-team-api/internal/modules/core/auth/repository"
	"slam-team-api/pkg/jwt"
	"slam-team-api/pkg/utils"

	"gorm.io/gorm"
)

// ErrInvalidCredentials is returned when the email/password pair does not match.
var ErrInvalidCredentials = errors.New("invalid email or password")

// AuthService composes the user repository with the JWT manager.
type AuthService struct {
	users  *repository.UserRepository
	jwtMgr *jwt.Manager
}

func NewAuthService(users *repository.UserRepository, jwtMgr *jwt.Manager) *AuthService {
	return &AuthService{users: users, jwtMgr: jwtMgr}
}

// Login verifies credentials and returns a signed JWT plus the user profile.
func (s *AuthService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
	}
	token, err := s.jwtMgr.Issue(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{Token: token, User: toUserResponse(user)}, nil
}

func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email}
}
