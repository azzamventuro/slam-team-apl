// Package jwt issues and verifies HS256 JSON Web Tokens for the API.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the token payload issued and verified by this service.
//
// The role fields are filled during login from the user's primary role in
// user_role → mst_role. PermVersion is bumped whenever the role_permission
// matrix changes, so the frontend can detect stale cached permissions.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	// AnggotaID is users.anggota_id — the member record behind this account.
	// Handlers use it to resolve the `milik_sendiri` cakupan (rows owned by
	// the caller) without a users lookup. Nil when the account has no anggota.
	AnggotaID *int64 `json:"anggota_id,omitempty"`
	// RoleID is mst_role.id; RoleLevel is mst_role.level (0 = super admin,
	// 10 = admin, 20 = moderator, 30 = user — lower is more privileged).
	RoleID    int64 `json:"role_id,omitempty"`
	RoleLevel int   `json:"role_level,omitempty"`
	// IsSuper mirrors mst_role.is_super: bypasses every permission check.
	IsSuper bool `json:"is_super,omitempty"`
	// PermVersion is a monotonically increasing counter from mst_pengaturan
	// "rbac.perm_version". The frontend stores this and re-fetches /me/permissions
	// when it changes, so permission changes propagate without re-login.
	PermVersion int64 `json:"perm_version,omitempty"`
	jwt.RegisteredClaims
}

// Manager issues and verifies HS256 tokens with a fixed secret and TTL.
type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// TTL is the access-token lifetime, exposed so login responses can report
// expires_in without a second source of truth.
func (m *Manager) TTL() time.Duration { return m.ttl }

// New returns a token Manager. secret must match across all verifiers.
func New(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Issue signs a token for the given user without role claims. This is the
// legacy signature kept for backward compatibility during the transition.
// New callers should use IssueWithRole.
func (m *Manager) Issue(userID uint, email, name string) (string, error) {
	return m.IssueWithRole(userID, email, name, 0, 0, false, 0)
}

// IssueWithRole signs a token with the full set of RBAC claims. The role
// fields come from user_role → mst_role during login. permVersion is the
// current value of mst_pengaturan "rbac.perm_version" and lets the frontend
// detect stale permission caches.
func (m *Manager) IssueWithRole(userID uint, email, name string, roleID int64, roleLevel int, isSuper bool, permVersion int64) (string, error) {
	return m.IssueAccess(userID, email, name, nil, roleID, roleLevel, isSuper, permVersion)
}

// IssueAccess is IssueWithRole plus the anggota_id claim. It is the signature
// the auth module uses at login and refresh; the older helpers remain as thin
// wrappers so existing callers keep compiling.
func (m *Manager) IssueAccess(userID uint, email, name string, anggotaID *int64, roleID int64, roleLevel int, isSuper bool, permVersion int64) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:      userID,
		Email:       email,
		Name:        name,
		AnggotaID:   anggotaID,
		RoleID:      roleID,
		RoleLevel:   roleLevel,
		IsSuper:     isSuper,
		PermVersion: permVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Verify parses and validates a token, returning its claims.
func (m *Manager) Verify(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
