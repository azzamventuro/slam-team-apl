// Package service holds the auth module's business logic: credential checks
// with lockout, access/refresh token issuance, session rotation and revocation,
// and the identity reads behind /me and /me/permissions.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/middleware"
	"slam-team-api/internal/modules/core/auth/domain"
	"slam-team-api/internal/modules/core/auth/dto"
	"slam-team-api/internal/modules/core/auth/repository"
	hakaksesrepo "slam-team-api/internal/modules/core/hakakses/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/pkg/jwt"
	"slam-team-api/pkg/utils"
)

// Lockout policy: after maxGagalLogin consecutive failures the account is
// locked for lockDuration. Constants for now; a mst_pengaturan key can replace
// them later without touching the flow.
const (
	maxGagalLogin = 5
	lockDuration  = 15 * time.Minute

	// refreshTokenBytes is the entropy of the opaque refresh token (43 chars
	// once base64url-encoded, well inside varchar(255)).
	refreshTokenBytes = 32
)

// authError carries a client-facing message while unwrapping to an apperr
// sentinel, so response.FromError still picks the right status.
type authError struct {
	msg  string
	base error
}

func (e *authError) Error() string { return e.msg }
func (e *authError) Unwrap() error { return e.base }

// ErrInvalidCredentials is the ONLY error a failed credential check returns —
// unknown identifier and wrong password are indistinguishable to the client.
var ErrInvalidCredentials = &authError{msg: "username/email atau password salah", base: apperr.ErrUnauthorized}

// ErrInvalidSession is returned for a refresh token that is unknown, revoked,
// or expired.
var ErrInvalidSession = &authError{msg: "sesi tidak valid atau sudah berakhir", base: apperr.ErrUnauthorized}

// ClientInfo is the request fingerprint recorded on sesi_login and the audit
// row. The handler fills it from the Gin context; the service stays HTTP-free.
type ClientInfo struct {
	IP        string
	UserAgent string
}

// AuthService composes the user/session repository with the JWT manager and
// the RBAC collaborators. hakaksesRepo and permGuard may be nil in tests;
// login then falls back to users.role_id for the role claims.
type AuthService struct {
	users        *repository.UserRepository
	jwtMgr       *jwt.Manager
	hakaksesRepo *hakaksesrepo.HakaksesRepository
	permGuard    *middleware.PermGuard
	auditor      *audit.Writer
	refreshTTL   time.Duration
}

func NewAuthService(
	users *repository.UserRepository,
	jwtMgr *jwt.Manager,
	hakaksesRepo *hakaksesrepo.HakaksesRepository,
	permGuard *middleware.PermGuard,
	auditor *audit.Writer,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:        users,
		jwtMgr:       jwtMgr,
		hakaksesRepo: hakaksesRepo,
		permGuard:    permGuard,
		auditor:      auditor,
		refreshTTL:   refreshTTL,
	}
}

// Login verifies an identifier (username or email) + password, enforces the
// lockout policy, and on success opens a sesi_login row and returns the token
// pair with the user profile.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest, client ClientInfo) (*dto.AuthResponse, error) {
	now := time.Now().UTC()

	user, err := s.users.FindByUsernameOrEmail(ctx, req.Identifier)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.Locked(now) {
		sisa := time.Until(*user.TerkunciSampai).Round(time.Second)
		return nil, &authError{
			msg:  fmt.Sprintf("akun terkunci, coba lagi dalam %s", sisa),
			base: apperr.ErrForbidden,
		}
	}
	if !user.IsAktif {
		return nil, &authError{msg: "akun nonaktif", base: apperr.ErrForbidden}
	}

	if user.Password == "" || !utils.CheckPassword(user.Password, req.Password) {
		count, incErr := s.users.IncrementGagalLogin(ctx, user.ID)
		if incErr != nil {
			return nil, incErr
		}
		if count >= maxGagalLogin {
			if lockErr := s.users.LockUntil(ctx, user.ID, now.Add(lockDuration)); lockErr != nil {
				return nil, lockErr
			}
		}
		return nil, ErrInvalidCredentials
	}

	if err := s.users.ResetGagalLogin(ctx, user.ID); err != nil {
		return nil, err
	}
	if err := s.users.TouchLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}

	resp, err := s.issuePair(ctx, user, client, now)
	if err != nil {
		return nil, err
	}

	actor := user.ID
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor,
		Modul:       "auth",
		Aksi:        "login",
		ReffType:    "users",
		ReffID:      &actor,
		Ringkasan:   fmt.Sprintf("Login %s", user.Username),
		IPAddress:   client.IP,
		UserAgent:   client.UserAgent,
	})
	return resp, nil
}

// Refresh exchanges an active refresh token for a new token pair. The old
// session is revoked first (rotation), so a replayed token fails with 401.
func (s *AuthService) Refresh(ctx context.Context, req dto.RefreshRequest, client ClientInfo) (*dto.AuthResponse, error) {
	now := time.Now().UTC()

	sesi, err := s.users.FindActiveSesiByToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}

	user, err := s.users.FindByID(ctx, sesi.UserID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}
	if !user.IsAktif || user.Locked(now) {
		// A disabled or locked account must not be able to keep itself alive
		// through refresh; close the session it presented.
		_ = s.users.RevokeSesi(ctx, req.RefreshToken, now)
		return nil, ErrInvalidSession
	}

	// Revoke before issuing: if the token was already spent by a concurrent
	// request, exactly one caller wins the rotation.
	if err := s.users.RevokeSesi(ctx, req.RefreshToken, now); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, ErrInvalidSession
		}
		return nil, err
	}
	return s.issuePair(ctx, user, client, now)
}

// Logout revokes the session named by the refresh token, provided it belongs
// to the caller. An already-revoked or unknown token is a no-op success — the
// client's goal (no live session) is met either way.
func (s *AuthService) Logout(ctx context.Context, claims *jwt.Claims, req dto.LogoutRequest, client ClientInfo) error {
	now := time.Now().UTC()

	sesi, err := s.users.FindActiveSesiByToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil
		}
		return err
	}
	if sesi.UserID != int64(claims.UserID) {
		return fmt.Errorf("sesi milik pengguna lain: %w", apperr.ErrForbidden)
	}
	if err := s.users.RevokeSesi(ctx, req.RefreshToken, now); err != nil && !errors.Is(err, apperr.ErrNotFound) {
		return err
	}

	actor := int64(claims.UserID)
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor,
		Modul:       "auth",
		Aksi:        "logout",
		ReffType:    "sesi_login",
		ReffID:      &sesi.ID,
		Ringkasan:   "Logout",
		IPAddress:   client.IP,
		UserAgent:   client.UserAgent,
	})
	return nil
}

// Me returns the authenticated account with its role and linked anggota.
func (s *AuthService) Me(ctx context.Context, claims *jwt.Claims) (*dto.MeResponse, error) {
	user, err := s.users.FindByID(ctx, int64(claims.UserID))
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.ErrUnauthorized
		}
		return nil, err
	}
	role, permVersion := s.roleClaims(ctx, user)

	resp := &dto.MeResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		AnggotaID:   user.AnggotaID,
		Nama:        displayName(user),
		Timezone:    user.Timezone,
		IsAktif:     user.IsAktif,
		Role:        role,
		PermVersion: permVersion,
	}
	if user.AnggotaID != nil && user.AnggotaNamaLengkap != nil {
		resp.Anggota = &dto.MeAnggota{
			ID:          *user.AnggotaID,
			NamaLengkap: *user.AnggotaNamaLengkap,
			NoInduk:     user.AnggotaNoInduk,
			FotoUUID:    user.AnggotaFotoUUID,
			FotoURL:     fotoURL(user.AnggotaFotoUUID),
		}
	}
	return resp, nil
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

// ── helpers ──

// issuePair signs an access token, mints an opaque refresh token, and records
// the session. Shared by Login and Refresh.
func (s *AuthService) issuePair(ctx context.Context, user *domain.User, client ClientInfo, now time.Time) (*dto.AuthResponse, error) {
	role, permVersion := s.roleClaims(ctx, user)

	access, err := s.jwtMgr.IssueAccess(
		uint(user.ID), user.Email, displayName(user), user.AnggotaID,
		role.ID, role.Level, role.IsSuper, permVersion,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal issue token: %w", err)
	}

	refresh, err := newRefreshToken()
	if err != nil {
		return nil, err
	}
	sesi := &domain.SesiLogin{
		UserID:        user.ID,
		RefreshToken:  refresh,
		IPAddress:     nilIfEmpty(client.IP),
		UserAgent:     nilIfEmpty(client.UserAgent),
		BerlakuSampai: now.Add(s.refreshTTL),
		CreatedAt:     now,
	}
	if err := s.users.CreateSesi(ctx, sesi); err != nil {
		return nil, fmt.Errorf("gagal membuat sesi: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		TokenType:    "Bearer",
		ExpiresIn:    int(s.jwtMgr.TTL().Seconds()),
		User: dto.AuthUser{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			AnggotaID:   user.AnggotaID,
			Nama:        displayName(user),
			FotoURL:     fotoURL(user.AnggotaFotoUUID),
			Role:        role,
			PermVersion: permVersion,
		},
	}, nil
}

// roleClaims resolves the role that goes into the token. The RBAC primary role
// (user_role → mst_role) wins because that is what PermGuard evaluates; when
// no user_role row exists the users.role_id join is the fallback so a freshly
// created account still gets a usable token.
func (s *AuthService) roleClaims(ctx context.Context, user *domain.User) (dto.AuthRole, int64) {
	role := dto.AuthRole{
		ID:      user.RoleID,
		Nama:    user.RoleNama,
		Level:   user.RoleLevel,
		IsSuper: user.RoleIsSuper,
	}
	var permVersion int64
	if s.hakaksesRepo == nil {
		return role, permVersion
	}
	permVersion, _ = s.hakaksesRepo.GetPermVersion(ctx)
	if info, _ := s.hakaksesRepo.GetUserPrimaryRole(ctx, user.ID); info != nil {
		role.ID = info.RoleID
		role.Level = info.Level
		role.IsSuper = info.IsSuper
		if info.RoleID != user.RoleID {
			role.Nama = info.KodeRole
		}
	}
	return role, permVersion
}

// displayName is the anggota's full name, falling back to the username for an
// account not yet linked to a member.
func displayName(u *domain.User) string {
	if u.AnggotaNamaLengkap != nil && *u.AnggotaNamaLengkap != "" {
		return *u.AnggotaNamaLengkap
	}
	return u.Username
}

// fotoURL builds the authenticated file URL for a profile photo uuid.
func fotoURL(uuid *string) *string {
	if uuid == nil || *uuid == "" {
		return nil
	}
	u := "/api/v1/files/" + *uuid + "/low"
	return &u
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// newRefreshToken returns 32 bytes of crypto/rand entropy, base64url-encoded
// without padding so it is safe in JSON bodies and query strings.
func newRefreshToken() (string, error) {
	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("gagal membuat refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
