// Package jwt issues and verifies HS256 JSON Web Tokens for the API.
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the token payload issued and verified by this service.
//
// The role fields are what the authorisation middleware reads. They are
// omitempty because the scaffold auth module does not issue them yet; a token
// without them is treated as having no role at all (RoleLevel 0 + IsSuper
// false), which every role gate denies. The real auth module fills them from
// users.role_id.
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	// RoleID is mst_role.id; RoleLevel is mst_role.level (0 = super admin,
	// 10 = admin, 20 = moderator, 30 = user — lower is more privileged).
	RoleID    int64 `json:"role_id,omitempty"`
	RoleLevel int   `json:"role_level,omitempty"`
	// IsSuper mirrors mst_role.is_super: bypasses every permission check.
	IsSuper bool `json:"is_super,omitempty"`
	jwt.RegisteredClaims
}

// Manager issues and verifies HS256 tokens with a fixed secret and TTL.
type Manager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// New returns a token Manager. secret must match across all verifiers.
func New(secret, issuer string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Issue signs a token for the given user.
func (m *Manager) Issue(userID uint, email, name string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
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
