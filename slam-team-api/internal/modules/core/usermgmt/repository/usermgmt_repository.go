// Package repository provides GORM data access for the user-management module.
// All queries are scoped to a role-level band so that each surface (admin /
// moderator / user) only sees and manages accounts within its target level range.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/usermgmt/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// UsermgmtRepository wraps *gorm.DB for user-management operations.
type UsermgmtRepository struct {
	db *gorm.DB
}

// NewUsermgmtRepository binds a repository to the process-wide DB pool.
func NewUsermgmtRepository(db *gorm.DB) *UsermgmtRepository {
	return &UsermgmtRepository{db: db}
}

// DB returns the underlying *gorm.DB for direct queries (e.g. role lookup).
func (r *UsermgmtRepository) DB() *gorm.DB {
	return r.db
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted users.
func (r *UsermgmtRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.User{}).Where("users.is_deleted = false")
}

// listJoin applies LEFT JOINs for anggota, role, and foto data.
// The caller must ensure the mst_role table is not already joined under
// the alias "r" (List does INNER JOIN mst_role r for band filtering).
func (r *UsermgmtRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			users.*,
			a.nama_lengkap    AS anggota_nama_lengkap,
			a.no_induk        AS anggota_no_induk,
			a.instansi_id     AS anggota_instansi_id,
			fp.uuid::text     AS anggota_foto_uuid,
			r.nama            AS role_nama,
			r.level           AS role_level,
			r.is_super        AS role_is_super
		`).
		Joins("LEFT JOIN anggota a ON a.id = users.anggota_id AND a.is_deleted = false").
		Joins("LEFT JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Joins("LEFT JOIN mst_file fp ON fp.id = a.foto_profil_file_id AND fp.is_deleted = false")
}

// listJoinNoRole applies LEFT JOINs for anggota and foto only, without joining
// mst_role. Use this when the caller already joined mst_role under alias "r"
// (e.g. List does INNER JOIN mst_role r for band filtering). The role columns
// (role_nama, role_level, role_is_super) are resolved from the pre-existing "r".
func (r *UsermgmtRepository) listJoinNoRole(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			users.*,
			a.nama_lengkap    AS anggota_nama_lengkap,
			a.no_induk        AS anggota_no_induk,
			a.instansi_id     AS anggota_instansi_id,
			fp.uuid::text     AS anggota_foto_uuid,
			r.nama            AS role_nama,
			r.level           AS role_level,
			r.is_super        AS role_is_super
		`).
		Joins("LEFT JOIN anggota a ON a.id = users.anggota_id AND a.is_deleted = false").
		Joins("LEFT JOIN mst_file fp ON fp.id = a.foto_profil_file_id AND fp.is_deleted = false")
}

// ── CRUD ──

// Create inserts a User row within a transaction, also creating the user_role row.
// Uses Omit to exclude joined/virtual fields that don't exist in the users table.
func (r *UsermgmtRepository) Create(ctx context.Context, u *domain.User, roleID int64, actorID *int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Insert only real users columns — skip joined fields.
		if err := tx.Omit(
			"anggota_nama_lengkap", "anggota_no_induk", "anggota_instansi_id",
			"anggota_foto_uuid", "role_nama", "role_level", "role_is_super",
		).Create(u).Error; err != nil {
			return err
		}

		// Insert the mirror user_role row (is_utama = true).
		ur := domain.UserRole{
			UserID:  u.ID,
			RoleID:  roleID,
			IsUtama: true,
		}
		ur.CreatedBy = actorID
		if err := tx.Create(&ur).Error; err != nil {
			return fmt.Errorf("create user_role: %w", err)
		}
		return nil
	})
}

// Update persists all mutable fields of an existing user row.
// Uses Omit to skip joined/virtual columns that don't exist in the users table.
func (r *UsermgmtRepository) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Omit(
		"anggota_nama_lengkap", "anggota_no_induk", "anggota_instansi_id",
		"anggota_foto_uuid", "role_nama", "role_level", "role_is_super",
	).Save(u).Error
}

// UpdateRoleSync updates the user_role mirror to match the user's current role_id.
// It sets the old primary role's is_utama=false and the new one's is_utama=true.
func (r *UsermgmtRepository) UpdateRoleSync(ctx context.Context, userID, newRoleID int64, actorID *int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Demote old primary role.
		if err := tx.Model(&domain.UserRole{}).
			Where("user_id = ? AND is_utama = true", userID).
			Update("is_utama", false).Error; err != nil {
			return err
		}
		// Check if a row for the new role already exists.
		var count int64
		if err := tx.Model(&domain.UserRole{}).
			Where("user_id = ? AND role_id = ?", userID, newRoleID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			// Update existing row to be primary.
			return tx.Model(&domain.UserRole{}).
				Where("user_id = ? AND role_id = ?", userID, newRoleID).
				Updates(map[string]any{
					"is_utama":   true,
					"created_by": actorID,
				}).Error
		}
		// Insert new primary role.
		ur := domain.UserRole{
			UserID:  userID,
			RoleID:  newRoleID,
			IsUtama: true,
		}
		ur.CreatedBy = actorID
		return tx.Create(&ur).Error
	})
}

// SoftDelete marks a user as deleted.
func (r *UsermgmtRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one user by id with joined anggota and role fields.
func (r *UsermgmtRepository) FindByID(ctx context.Context, id int64) (*domain.User, error) {
	var u domain.User
	q := r.baseQuery().Where("users.id = ?", id)
	// listJoin adds LEFT JOIN mst_role r — no prior role join here.
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user: %w", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &u, nil
}

// List returns a paginated slice of users filtered by role-level band.
// bandMin and bandMax define the inclusive range of role.level to include.
func (r *UsermgmtRepository) List(ctx context.Context, q ListQuery, bandMin, bandMax int) ([]domain.User, int64, error) {
	base := r.baseQuery().
		Joins("INNER JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Where("r.level >= ? AND r.level <= ?", bandMin, bandMax)

	// Filter by search query (ILIKE on username, email, anggota nama_lengkap).
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(users.username) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(a.nama_lengkap) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by is_aktif.
	if q.IsAktif != nil {
		base = base.Where("users.is_aktif = ?", *q.IsAktif)
	}

	// Filter by instansi_id (via anggota).
	if q.InstansiID != nil {
		base = base.Where("a.instansi_id = ?", *q.InstansiID)
	}

	// Count total rows (without joins for accuracy).
	var total int64
	countQ := r.db.Model(&domain.User{}).
		Where("users.is_deleted = false").
		Joins("INNER JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Where("r.level >= ? AND r.level <= ?", bandMin, bandMax)
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		countQ = countQ.Joins("LEFT JOIN anggota a ON a.id = users.anggota_id AND a.is_deleted = false").
			Where(
				"(LOWER(users.username) LIKE ? OR LOWER(users.email) LIKE ? OR LOWER(a.nama_lengkap) LIKE ?)",
				like, like, like,
			)
	}
	if q.IsAktif != nil {
		countQ = countQ.Where("users.is_aktif = ?", *q.IsAktif)
	}
	if q.InstansiID != nil {
		if q.Q == "" {
			countQ = countQ.Joins("LEFT JOIN anggota a ON a.id = users.anggota_id AND a.is_deleted = false")
		}
		countQ = countQ.Where("a.instansi_id = ?", *q.InstansiID)
	}
	if err := countQ.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort.
	base = r.applySort(base, q.Sort)

	// Paginated query with joins — listJoin adds LEFT JOIN mst_role r,
	// but List already has INNER JOIN mst_role r, so we use a subquery
	// approach: remove the prior role join alias conflict by switching
	// listJoin to use alias "rl" for role when called from List.
	//
	// Actually simpler: re-alias the role columns via the already-joined "r".
	paged := r.listJoinNoRole(base).Offset(q.Offset).Limit(q.Limit)

	var items []domain.User
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ── Validation helpers ──

// ExistsUsername checks whether a username already exists among non-deleted users,
// optionally excluding a given id.
func (r *UsermgmtRepository) ExistsUsername(ctx context.Context, username string, exceptID int64) (bool, error) {
	var count int64
	q := r.baseQuery().Where("LOWER(users.username) = LOWER(?)", username)
	if exceptID > 0 {
		q = q.Where("users.id <> ?", exceptID)
	}
	if err := q.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ExistsEmail checks whether an email already exists among non-deleted users,
// optionally excluding a given id.
func (r *UsermgmtRepository) ExistsEmail(ctx context.Context, email string, exceptID int64) (bool, error) {
	var count int64
	q := r.baseQuery().Where("LOWER(users.email) = LOWER(?)", email)
	if exceptID > 0 {
		q = q.Where("users.id <> ?", exceptID)
	}
	if err := q.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// AnggotaExists checks whether a non-deleted anggota row exists.
func (r *UsermgmtRepository) AnggotaExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("anggota").
		Where("id = ? AND is_deleted = false", id).
		Count(&count).Error
	return count > 0, err
}

// AnggotaHasAccount checks whether an anggota already has a non-deleted user account.
func (r *UsermgmtRepository) AnggotaHasAccount(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.baseQuery().Where("users.anggota_id = ?", anggotaID).Count(&count).Error
	return count > 0, err
}

// RoleExists checks whether a non-deleted, active role exists.
func (r *UsermgmtRepository) RoleExists(ctx context.Context, roleID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Role{}).
		Where("id = ? AND is_deleted = false AND is_aktif = true", roleID).
		Count(&count).Error
	return count > 0, err
}

// CountSuperAdmin returns the count of active, non-deleted super admin accounts.
func (r *UsermgmtRepository) CountSuperAdmin(ctx context.Context) (int64, error) {
	var count int64
	err := r.baseQuery().
		Joins("INNER JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Where("r.is_super = true AND users.is_deleted = false").
		WithContext(ctx).Count(&count).Error
	return count, err
}

// IsSuperAdmin checks if a specific user is a super admin.
func (r *UsermgmtRepository) IsSuperAdmin(ctx context.Context, userID int64) (bool, error) {
	var count int64
	err := r.baseQuery().
		Joins("INNER JOIN mst_role r ON r.id = users.role_id AND r.is_deleted = false").
		Where("users.id = ? AND r.is_super = true", userID).
		WithContext(ctx).Count(&count).Error
	return count > 0, err
}

// RolesByBand returns roles within the given level range that the actor can manage.
// Filters: level within band, level > actorLevel (anti-escalation), is_super=false,
// is_aktif=true, is_deleted=false.
func (r *UsermgmtRepository) RolesByBand(ctx context.Context, bandMin, bandMax, actorLevel int) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).
		Where("level >= ? AND level <= ? AND level > ? AND is_super = false AND is_aktif = true AND is_deleted = false",
			bandMin, bandMax, actorLevel).
		Order("level ASC").
		Find(&roles).Error
	return roles, err
}

// AnggotaWithoutAccount returns anggota who don't have an active user account,
// optionally filtered by search query and instansi_id.
func (r *UsermgmtRepository) AnggotaWithoutAccount(ctx context.Context, q string, instansiID *int64, limit int) ([]domain.AnggotaLite, error) {
	var items []domain.AnggotaLite
	db := r.db.WithContext(ctx).
		Table("anggota a").
		Select("a.id, a.nama_lengkap, COALESCE(a.no_induk, '') AS no_induk, a.instansi_id, COALESCE(fp.uuid::text, '') AS foto_uuid").
		Joins("LEFT JOIN mst_file fp ON fp.id = a.foto_profil_file_id AND fp.is_deleted = false").
		Where("a.is_deleted = false AND a.id NOT IN (SELECT u.anggota_id FROM users u WHERE u.is_deleted = false)")

	if q != "" {
		like := "%" + strings.ToLower(q) + "%"
		db = db.Where("(LOWER(a.nama_lengkap) LIKE ? OR LOWER(a.no_induk) LIKE ?)", like, like)
	}
	if instansiID != nil {
		db = db.Where("a.instansi_id = ?", *instansiID)
	}

	if limit <= 0 {
		limit = 20
	}
	if err := db.Order("a.nama_lengkap ASC").Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// ── Sort ──

var allowedUserSortColumns = map[string]bool{
	"username": true, "email": true, "created_at": true, "login_terakhir": true,
}

func (r *UsermgmtRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("users.created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedUserSortColumns[col] {
		return q.Order("users.created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("users.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List.
type ListQuery struct {
	Q          string
	IsAktif    *bool
	InstansiID *int64
	Sort       string
	Offset     int
	Limit      int
}
