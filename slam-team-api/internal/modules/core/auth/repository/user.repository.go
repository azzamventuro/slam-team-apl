// Package repository is the auth module's GORM data-access layer over users
// and sesi_login.
package repository

import (
	"context"
	"errors"
	"time"

	"slam-team-api/internal/modules/core/auth/domain"
	"slam-team-api/internal/shared/apperr"

	"gorm.io/gorm"
)

// UserRepository reads and writes users and their login sessions.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// profileQuery is the users row plus the joined anggota / foto / role summary
// every auth read needs (login response, /me). The joins are LEFT so an
// account whose anggota or role row is missing still authenticates.
func (r *UserRepository) profileQuery(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Select(`
			users.*,
			a.nama_lengkap    AS anggota_nama_lengkap,
			a.no_induk        AS anggota_no_induk,
			fp.uuid::text     AS anggota_foto_uuid,
			r.nama            AS role_nama,
			r.level           AS role_level,
			r.is_super        AS role_is_super
		`).
		Joins("LEFT JOIN anggota a ON a.id = users.anggota_id AND a.is_deleted = false").
		Joins("LEFT JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Joins("LEFT JOIN mst_file fp ON fp.id = a.foto_profil_file_id AND fp.is_deleted = false").
		Where("users.is_deleted = false")
}

// FindByUsernameOrEmail resolves a login identifier against username OR email
// in a single query (both are unique among non-deleted rows). Returns
// apperr.ErrNotFound when nothing matches; the service turns that into the
// generic credentials error.
func (r *UserRepository) FindByUsernameOrEmail(ctx context.Context, identifier string) (*domain.User, error) {
	var user domain.User
	err := r.profileQuery(ctx).
		Where("(users.username = ? OR users.email = ?)", identifier, identifier).
		Limit(1).
		Scan(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, apperr.ErrNotFound
	}
	return &user, nil
}

// FindByID returns one non-deleted user with the joined profile summary.
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.profileQuery(ctx).Where("users.id = ?", id).Limit(1).Scan(&user).Error
	if err != nil {
		return nil, err
	}
	if user.ID == 0 {
		return nil, apperr.ErrNotFound
	}
	return &user, nil
}

// TouchLogin stamps login_terakhir after a successful authentication.
func (r *UserRepository) TouchLogin(ctx context.Context, userID int64, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("login_terakhir", at).Error
}

// IncrementGagalLogin bumps the consecutive-failure counter and returns the
// new value, so the caller can decide whether the lockout threshold is hit
// without a second round-trip.
func (r *UserRepository) IncrementGagalLogin(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.WithContext(ctx).Raw(`
		UPDATE users SET gagal_login = gagal_login + 1
		WHERE id = ? RETURNING gagal_login
	`, userID).Scan(&count).Error
	return count, err
}

// ResetGagalLogin clears the failure counter and any lockout after success.
func (r *UserRepository) ResetGagalLogin(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{"gagal_login": 0, "terkunci_sampai": nil}).Error
}

// LockUntil sets terkunci_sampai; login is refused until that instant.
func (r *UserRepository) LockUntil(ctx context.Context, userID int64, until time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("terkunci_sampai", until).Error
}

// ── sesi_login ──

// CreateSesi inserts one session row (login or refresh rotation).
func (r *UserRepository) CreateSesi(ctx context.Context, s *domain.SesiLogin) error {
	return r.db.WithContext(ctx).Create(s).Error
}

// FindActiveSesiByToken returns the session for a refresh token that is
// neither revoked nor expired. A revoked, expired, or unknown token all map to
// apperr.ErrNotFound — the service does not distinguish them to the client.
func (r *UserRepository) FindActiveSesiByToken(ctx context.Context, token string) (*domain.SesiLogin, error) {
	var s domain.SesiLogin
	err := r.db.WithContext(ctx).
		Where("refresh_token = ? AND dicabut_pada IS NULL AND berlaku_sampai > now()", token).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.ErrNotFound
		}
		return nil, err
	}
	return &s, nil
}

// RevokeSesi marks one session revoked. Already-revoked rows are left alone so
// the original revocation time survives. Returns apperr.ErrNotFound when no
// active session had that token.
func (r *UserRepository) RevokeSesi(ctx context.Context, token string, at time.Time) error {
	res := r.db.WithContext(ctx).Model(&domain.SesiLogin{}).
		Where("refresh_token = ? AND dicabut_pada IS NULL", token).
		Update("dicabut_pada", at)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

// RevokeAllSesi revokes every active session of a user (forced logout,
// password change). Returns the number of sessions revoked.
func (r *UserRepository) RevokeAllSesi(ctx context.Context, userID int64, at time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.SesiLogin{}).
		Where("user_id = ? AND dicabut_pada IS NULL", userID).
		Update("dicabut_pada", at)
	return res.RowsAffected, res.Error
}
