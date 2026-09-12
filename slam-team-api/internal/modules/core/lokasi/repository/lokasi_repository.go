// Package repository provides GORM data access for the lokasi module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/lokasi/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// LokasiRepository wraps *gorm.DB for mst_lokasi operations.
type LokasiRepository struct {
	db *gorm.DB
}

// NewLokasiRepository binds a repository to the process-wide DB pool.
func NewLokasiRepository(db *gorm.DB) *LokasiRepository {
	return &LokasiRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows, used by every read.
//
// is_deleted is qualified to mst_lokasi (rather than using the shared
// model.NotDeleted scope) because every read here joins mst_file, which also
// carries an is_deleted column — an unqualified WHERE is_deleted = false is
// ambiguous once that join is in the FROM clause.
func (r *LokasiRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Lokasi{}).Where("mst_lokasi.is_deleted = false")
}

// listJoin applies the LEFT JOIN for the foto UUID. Call before Find/Count.
func (r *LokasiRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			mst_lokasi.*,
			f.uuid AS foto_uuid
		`).
		Joins("LEFT JOIN mst_file f ON f.id = mst_lokasi.foto_file_id AND f.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *LokasiRepository) Create(ctx context.Context, l *domain.Lokasi) error {
	return r.db.WithContext(ctx).Omit("foto_uuid").Create(l).Error
}

// Update persists all mutable fields of an existing row.
func (r *LokasiRepository) Update(ctx context.Context, l *domain.Lokasi) error {
	return r.db.WithContext(ctx).Omit("foto_uuid").Save(l).Error
}

// SoftDelete marks a row as deleted. It does NOT check foreign-key guards —
// the caller (service) is responsible for that.
func (r *LokasiRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Lokasi{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one lokasi by id, with foto UUID.
func (r *LokasiRepository) FindByID(ctx context.Context, id int64) (*domain.Lokasi, error) {
	var l domain.Lokasi
	q := r.baseQuery().Where("mst_lokasi.id = ?", id)
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&l).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: lokasi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &l, nil
}

// List returns a paginated slice of lokasi with foto UUID.
func (r *LokasiRepository) List(ctx context.Context, q ListQuery) ([]domain.Lokasi, int64, error) {
	base := r.baseQuery()

	// Filter by search query.
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(mst_lokasi.kode) LIKE ? OR LOWER(mst_lokasi.nama) LIKE ? OR LOWER(mst_lokasi.alamat) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by is_aktif.
	if q.IsAktif != nil {
		base = base.Where("mst_lokasi.is_aktif = ?", *q.IsAktif)
	}

	// Count total rows (without joins).
	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort.
	base = r.applySort(base, q.Sort)

	// Paginated query with joins.
	paged := r.listJoin(base).Offset(q.Offset).Limit(q.Limit)

	var items []domain.Lokasi
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ExistsKode checks whether a kode already exists among non-deleted rows,
// optionally excluding a given id (for update uniqueness).
func (r *LokasiRepository) ExistsKode(ctx context.Context, kode string, exceptID int64) (bool, error) {
	var count int64
	q := r.baseQuery().Where("kode = ?", kode)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	if err := q.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// FileExists checks whether a mst_file row exists and is not soft-deleted.
func (r *LokasiRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"kode": true, "nama": true, "created_at": true,
}

func (r *LokasiRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("mst_lokasi.created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("mst_lokasi.created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("mst_lokasi.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List, bridging from the DTO.
// It is exported so the service layer can construct it.
type ListQuery struct {
	Q       string
	IsAktif *bool
	Sort    string
	Offset  int
	Limit   int
}
