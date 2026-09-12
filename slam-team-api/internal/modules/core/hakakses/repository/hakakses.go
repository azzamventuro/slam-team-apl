// Package repository provides GORM queries for the hakakses (RBAC) module.
package repository

import (
	"context"

	"slam-team-api/internal/modules/core/hakakses/domain"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type HakaksesRepository struct {
	db *gorm.DB
}

func NewHakaksesRepository(db *gorm.DB) *HakaksesRepository {
	return &HakaksesRepository{db: db}
}

// ── Role CRUD ──

func (r *HakaksesRepository) CreateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *HakaksesRepository) GetRoleByID(ctx context.Context, id int64) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).
		Scopes(model.NotDeleted).
		First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *HakaksesRepository) GetRoleByKode(ctx context.Context, kode string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).
		Scopes(model.NotDeleted).
		Where("kode = ?", kode).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *HakaksesRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	var roles []domain.Role
	err := r.db.WithContext(ctx).
		Scopes(model.NotDeleted).
		Order("level ASC").
		Find(&roles).Error
	return roles, err
}

func (r *HakaksesRepository) UpdateRole(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *HakaksesRepository) SoftDeleteRole(ctx context.Context, id int64, deletedBy int64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Role{}).
		Where("id = ? AND is_sistem = false", id).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_by": deletedBy,
			"is_aktif":   false,
		}).Error
}

func (r *HakaksesRepository) CountUsersByRoleID(ctx context.Context, roleID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return count, err
}

// ── Permission catalogue ──

func (r *HakaksesRepository) ListModulWithPermissions(ctx context.Context) ([]domain.Modul, error) {
	var moduls []domain.Modul
	err := r.db.WithContext(ctx).
		Where("is_aktif = true").
		Order("urutan ASC").
		Find(&moduls).Error
	return moduls, err
}

func (r *HakaksesRepository) ListPermissionsByModulID(ctx context.Context, modulID int64) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).
		Where("modul_id = ?", modulID).
		Order("aksi ASC").
		Find(&perms).Error
	return perms, err
}

func (r *HakaksesRepository) ListAllPermissions(ctx context.Context) ([]domain.Permission, error) {
	var perms []domain.Permission
	err := r.db.WithContext(ctx).Order("kode ASC").Find(&perms).Error
	return perms, err
}

func (r *HakaksesRepository) GetPermissionByKode(ctx context.Context, kode string) (*domain.Permission, error) {
	var perm domain.Permission
	err := r.db.WithContext(ctx).Where("kode = ?", kode).First(&perm).Error
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

// ── Matrix (role_permission) ──

// GetRolePermissions returns all permission_ids and cakupan for a role.
func (r *HakaksesRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]domain.RolePermission, error) {
	var rows []domain.RolePermission
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&rows).Error
	return rows, err
}

// ReplaceRolePermissions atomically replaces all role_permission rows for a role.
// It deletes existing rows and inserts the new set in a single transaction.
func (r *HakaksesRepository) ReplaceRolePermissions(ctx context.Context, roleID int64, grants []domain.RolePermission) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing.
		if err := tx.Where("role_id = ?", roleID).Delete(&domain.RolePermission{}).Error; err != nil {
			return err
		}
		// Insert new (batch).
		if len(grants) == 0 {
			return nil
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&grants).Error
	})
}

// ── PermGuard support ──

// EffectivePermissions returns the set of permission.kode values effective for
// a given role_id, combining role_permission + cakupan.
func (r *HakaksesRepository) EffectivePermissions(ctx context.Context, roleID int64) ([]EffectivePerm, error) {
	var rows []EffectivePerm
	err := r.db.WithContext(ctx).
		Table("role_permission rp").
		Joins("JOIN mst_permission p ON p.id = rp.permission_id").
		Joins("JOIN mst_modul m ON m.id = p.modul_id AND m.is_aktif = true").
		Where("rp.role_id = ?", roleID).
		Select("p.kode, rp.cakupan").
		Scan(&rows).Error
	return rows, err
}

// EffectivePerm is a lightweight projection returned by EffectivePermissions.
type EffectivePerm struct {
	Kode    string `gorm:"column:kode"`
	Cakupan string `gorm:"column:cakupan"`
}

// ── Perm version ──

// GetPermVersion reads the current perm_version counter from mst_pengaturan.
func (r *HakaksesRepository) GetPermVersion(ctx context.Context) (int64, error) {
	// The scan target must be exported: GORM ignores unexported struct fields,
	// which used to leave the value empty and every token at perm_version 0.
	var nilai string
	err := r.db.WithContext(ctx).
		Raw("SELECT nilai FROM mst_pengaturan WHERE kunci = 'rbac.perm_version'").
		Scan(&nilai).Error
	if err != nil {
		return 0, err
	}
	// Parse as int64.
	var v int64
	for _, c := range nilai {
		if c >= '0' && c <= '9' {
			v = v*10 + int64(c-'0')
		}
	}
	return v, nil
}

// BumpPermVersion increments the perm_version counter (called after any matrix change).
func (r *HakaksesRepository) BumpPermVersion(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Exec("UPDATE mst_pengaturan SET nilai = CAST((CAST(nilai AS bigint) + 1) AS varchar) WHERE kunci = 'rbac.perm_version'").
		Error
}

// ── User role lookup (used by auth module) ──

// UserRoleInfo holds the joined role data needed for JWT claims.
type UserRoleInfo struct {
	RoleID    int64
	Level     int
	IsSuper   bool
	KodeRole  string
}

// GetUserPrimaryRole returns the user's primary role (is_utama = true, or the
// lowest-level active role). Returns nil when the user has no role assigned.
func (r *HakaksesRepository) GetUserPrimaryRole(ctx context.Context, userID int64) (*UserRoleInfo, error) {
	var info UserRoleInfo
	err := r.db.WithContext(ctx).
		Table("user_role ur").
		Joins("JOIN mst_role r ON r.id = ur.role_id AND r.is_aktif = true AND r.is_deleted = false").
		Where("ur.user_id = ?", userID).
		Order("r.level ASC, ur.is_utama DESC").
		Select("r.id AS role_id, r.level, r.is_super, r.kode AS kode_role").
		Limit(1).
		Scan(&info).Error
	if err != nil {
		return nil, err
	}
	if info.RoleID == 0 {
		return nil, nil
	}
	return &info, nil
}
