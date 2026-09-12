// Package service holds the auth module's business logic.
package service

import (
	"context"
	"errors"
	"fmt"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/domain"
	"slam-team-api/internal/modules/core/auth/dto"
	"slam-team-api/internal/modules/core/auth/repository"
	hakaksesrepo "slam-team-api/internal/modules/core/hakakses/repository"
	"slam-team-api/pkg/jwt"
	"slam-team-api/pkg/utils"

	"gorm.io/gorm"
)

// ErrInvalidCredentials is returned when the email/password pair does not match.
var ErrInvalidCredentials = errors.New("invalid email or password")

// AuthService composes the user repository with the JWT manager and RBAC
// dependencies. hakaksesRepo and permGuard may be nil during the transition
// period before the hakakses module is wired.
type AuthService struct {
	users       *repository.UserRepository
	jwtMgr      *jwt.Manager
	hakaksesRepo *hakaksesrepo.HakaksesRepository
	permGuard    *middleware.PermGuard
}

func NewAuthService(
	users *repository.UserRepository,
	jwtMgr *jwt.Manager,
	hakaksesRepo *hakaksesrepo.HakaksesRepository,
	permGuard *middleware.PermGuard,
) *AuthService {
	return &AuthService{
		users:        users,
		jwtMgr:       jwtMgr,
		hakaksesRepo: hakaksesRepo,
		permGuard:    permGuard,
	}
}

// Login verifies credentials and returns a signed JWT plus the user profile.
// When RBAC is available, the token includes role claims and perm_version.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
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

	// Build JWT — with or without RBAC claims.
	var token string
	if s.hakaksesRepo != nil {
		roleInfo, _ := s.hakaksesRepo.GetUserPrimaryRole(ctx, int64(user.ID))
		permVersion, _ := s.hakaksesRepo.GetPermVersion(ctx)
		var roleID int64
		var roleLevel int
		var isSuper bool
		if roleInfo != nil {
			roleID = roleInfo.RoleID
			roleLevel = roleInfo.Level
			isSuper = roleInfo.IsSuper
		}
		token, err = s.jwtMgr.IssueWithRole(
			user.ID, user.Email, user.Name,
			roleID, roleLevel, isSuper, permVersion,
		)
	} else {
		token, err = s.jwtMgr.Issue(user.ID, user.Email, user.Name)
	}
	if err != nil {
		return nil, fmt.Errorf("gagal issue token: %w", err)
	}
	return &dto.LoginResponse{Token: token, User: toUserResponse(user)}, nil
}

// MyPermissions returns the effective permission set for the authenticated user.
func (s *AuthService) MyPermissions(ctx context.Context, claims *jwt.Claims) (*dto.MyPermissionsResp, error) {
	if claims.IsSuper {
		permVersion, _ := s.hakaksesRepo.GetPermVersion(ctx)
		return &dto.MyPermissionsResp{
			Permissions: []string{"*"}, // super admin: all
			PermVersion: permVersion,
			IsSuper:     true,
		}, nil
	}
	perms, err := s.permGuard.Effective(ctx, claims.RoleID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(perms))
	for kode := range perms {
		codes = append(codes, kode)
	}
	permVersion, _ := s.hakaksesRepo.GetPermVersion(ctx)
	return &dto.MyPermissionsResp{
		Permissions: codes,
		PermVersion: permVersion,
		IsSuper:     false,
	}, nil
}

func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{ID: u.ID, Name: u.Name, Email: u.Email}
}
