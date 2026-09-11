// Package repository provides GORM data access for the dokumen module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/dokumen/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// DokumenRepository wraps *gorm.DB for mst_dokumen operations.
type DokumenRepository struct {
	db *gorm.DB
}

// NewDokumenRepository binds a repository to the process-wide DB pool.
func NewDokumenRepository(db *gorm.DB) *DokumenRepository {
	return &DokumenRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *DokumenRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Dokumen{}).Scopes(model.NotDeleted)
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *DokumenRepository) Create(ctx context.Context, d *domain.Dokumen) error {
	return r.db.WithContext(ctx).Omit("File").Create(d).Error
}

// Update persists all mutable fields of an existing row.
func (r *DokumenRepository) Update(ctx context.Context, d *domain.Dokumen) error {
	return r.db.WithContext(ctx).Omit("File").Save(d).Error
}

// SoftDelete marks a row as deleted.
func (r *DokumenRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Dokumen{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one document by id, with the attached file preloaded.
func (r *DokumenRepository) FindByID(ctx context.Context, id int64) (*domain.Dokumen, error) {
	var d domain.Dokumen
	if err := r.baseQuery().WithContext(ctx).
		Preload("File").
		Where("id = ?", id).
		First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: dokumen", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &d, nil
}

// List returns a paginated slice of documents with file preloaded.
func (r *DokumenRepository) List(ctx context.Context, q ListQuery) ([]domain.Dokumen, int64, error) {
	base := r.baseQuery()

	// Filter by search query (kode + keterangan).
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(kode) LIKE ? OR LOWER(keterangan) LIKE ?)",
			like, like,
		)
	}

	// Filter by reff_type.
	if q.ReffType != "" {
		base = base.Where("reff_type = ?", q.ReffType)
	}

	// Filter by reff_id.
	if q.ReffID != nil {
		base = base.Where("reff_id = ?", *q.ReffID)
	}

	// Filter by jenis.
	if q.Jenis != nil {
		base = base.Where("jenis = ?", *q.Jenis)
	}

	// Filter by tipe.
	if q.Tipe != "" {
		base = base.Where("LOWER(tipe) = ?", strings.ToLower(q.Tipe))
	}

	// Filter by format.
	if q.Format != "" {
		base = base.Where("LOWER(format) = ?", strings.ToLower(q.Format))
	}

	// Count total rows.
	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort.
	base = r.applySort(base, q.Sort)

	// Paginated query with File preload.
	paged := base.Offset(q.Offset).Limit(q.Limit)

	var items []domain.Dokumen
	if err := paged.WithContext(ctx).Preload("File").Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"created_at": true, "modified_at": true, "jenis": true, "tipe": true,
}

func (r *DokumenRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
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
	Q        string
	ReffType string
	ReffID   *int64
	Jenis    *int
	Tipe     string
	Format   string
	Sort     string
	Offset   int
	Limit    int
}
