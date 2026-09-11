// Package repository provides GORM data access for the medsos module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/medsos/domain"
	"slam-team-api/internal/modules/core/medsos/dto"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// MedsosRepository wraps *gorm.DB for mst_medsos operations.
type MedsosRepository struct {
	db *gorm.DB
}

// NewMedsosRepository binds a repository to the process-wide DB pool.
func NewMedsosRepository(db *gorm.DB) *MedsosRepository {
	return &MedsosRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *MedsosRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Medsos{}).Scopes(model.NotDeleted)
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *MedsosRepository) Create(ctx context.Context, m *domain.Medsos) error {
	return r.db.WithContext(ctx).Create(m).Error
}

// Update persists all mutable fields of an existing row.
func (r *MedsosRepository) Update(ctx context.Context, m *domain.Medsos) error {
	return r.db.WithContext(ctx).Save(m).Error
}

// SoftDelete marks a row as deleted.
func (r *MedsosRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Medsos{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one medsos by id.
func (r *MedsosRepository) FindByID(ctx context.Context, id int64) (*domain.Medsos, error) {
	var m domain.Medsos
	if err := r.baseQuery().WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: medsos", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &m, nil
}

// FindByIDScoped returns one medsos by id, optionally scoped by anggota_id (milik_sendiri).
func (r *MedsosRepository) FindByIDScoped(ctx context.Context, id int64, anggotaID *int64) (*domain.Medsos, error) {
	var m domain.Medsos
	q := r.baseQuery().Where("id = ?", id)
	if anggotaID != nil {
		q = q.Where("anggota_id = ?", *anggotaID)
	}
	if err := q.WithContext(ctx).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: medsos", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &m, nil
}

// List returns a paginated slice of medsos.
func (r *MedsosRepository) List(ctx context.Context, q ListQuery, anggotaScope *int64) ([]domain.Medsos, int64, error) {
	base := r.baseQuery()

	// Scope: milik_sendiri filter.
	if anggotaScope != nil {
		base = base.Where("anggota_id = ?", *anggotaScope)
	}

	// Filter by search query.
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(jenis_medsos) LIKE ? OR LOWER(konten_medsos) LIKE ? OR LOWER(kode) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by anggota_id (explicit query param, distinct from scope).
	if q.AnggotaID != nil {
		base = base.Where("anggota_id = ?", *q.AnggotaID)
	}

	// Filter by jenis_medsos.
	if q.JenisMedsos != "" {
		base = base.Where("LOWER(jenis_medsos) = ?", strings.ToLower(q.JenisMedsos))
	}

	// Count total rows.
	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort.
	base = r.applySort(base, q.Sort)

	// Paginated query.
	paged := base.Offset(q.Offset).Limit(q.Limit)

	var items []domain.Medsos
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// AnggotaExists checks whether an anggota row exists and is not soft-deleted.
func (r *MedsosRepository) AnggotaExists(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("anggota").
		Where("id = ? AND is_deleted = false", anggotaID).
		Count(&count).Error
	return count > 0, err
}

// AnggotaTersedia returns all active anggota for dropdowns.
func (r *MedsosRepository) AnggotaTersedia(ctx context.Context, items *[]dto.AnggotaOpsi) error {
	return r.db.WithContext(ctx).
		Table("anggota").
		Select("id, nama_lengkap").
		Where("is_deleted = false").
		Order("nama_lengkap ASC").
		Scan(items).Error
}

// GetAnggotaIDByUserID looks up the anggota_id for a given user_id.
func (r *MedsosRepository) GetAnggotaIDByUserID(ctx context.Context, userID int64) (*int64, error) {
	var result struct {
		AnggotaID *int64 `gorm:"column:anggota_id"`
	}
	err := r.db.WithContext(ctx).Table("users").
		Select("anggota_id").
		Where("id = ? AND is_deleted = false", userID).
		Scan(&result).Error
	if err != nil {
		return nil, err
	}
	return result.AnggotaID, nil
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"created_at": true, "jenis_medsos": true, "tipe": true,
}

func (r *MedsosRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("%s %s", col, dir))
}

// ListQuery is the internal query shape used by List, bridging from the DTO.
type ListQuery struct {
	Q           string
	AnggotaID   *int64
	JenisMedsos string
	Sort        string
	Offset      int
	Limit       int
}
