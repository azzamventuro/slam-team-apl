// Package repository provides GORM data access for the instansi module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/instansi/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// InstansiRepository wraps *gorm.DB for mst_instansi operations.
type InstansiRepository struct {
	db *gorm.DB
}

// NewInstansiRepository binds a repository to the process-wide DB pool.
func NewInstansiRepository(db *gorm.DB) *InstansiRepository {
	return &InstansiRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows, used by every read.
func (r *InstansiRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Instansi{}).Scopes(model.NotDeleted)
}

// listJoin applies the two LEFT JOINs (logo UUIDs) and the sub-select
// jumlah_anggota to a query. Call before Find/Count.
func (r *InstansiRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			mst_instansi.*,
			lu.uuid   AS logo_utama_uuid,
			lt.uuid   AS logo_tambahan_uuid,
			(SELECT COUNT(*) FROM anggota a
			 WHERE a.instansi_id = mst_instansi.id AND a.is_deleted = false
			) AS jumlah_anggota
		`).
		LeftJoin("mst_file lu ON lu.id = mst_instansi.logo_utama_file_id AND lu.is_deleted = false").
		LeftJoin("mst_file lt ON lt.id = mst_instansi.logo_tambahan_file_id AND lt.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *InstansiRepository) Create(ctx context.Context, inst *domain.Instansi) error {
	return r.db.WithContext(ctx).Create(inst).Error
}

// Update persists all mutable fields of an existing row.
func (r *InstansiRepository) Update(ctx context.Context, inst *domain.Instansi) error {
	return r.db.WithContext(ctx).Save(inst).Error
}

// SoftDelete marks a row as deleted. It does NOT check foreign-key guards —
// the caller (service) is responsible for that.
func (r *InstansiRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Instansi{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one instansi by id, with logo UUIDs and jumlah_anggota.
func (r *InstansiRepository) FindByID(ctx context.Context, id int64) (*domain.Instansi, error) {
	var inst domain.Instansi
	q := r.baseQuery().Where("mst_instansi.id = ?", id)
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&inst).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: instansi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &inst, nil
}

// List returns a paginated slice of instansi with logo UUIDs and jumlah_anggota.
func (r *InstansiRepository) List(ctx context.Context, q ListQuery) ([]domain.Instansi, int64, error) {
	base := r.baseQuery()

	// Filter by search query.
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(mst_instansi.kode) LIKE ? OR LOWER(mst_instansi.nama) LIKE ? OR LOWER(mst_instansi.nama_club) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by status.
	if q.Status != nil {
		base = base.Where("mst_instansi.status = ?", *q.Status)
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

	var items []domain.Instansi
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ExistsKode checks whether a kode already exists among non-deleted rows,
// optionally excluding a given id (for update uniqueness).
func (r *InstansiRepository) ExistsKode(ctx context.Context, kode string, exceptID int64) (bool, error) {
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

// CountAnggotaAktif returns the number of active anggota pointing at this instansi.
func (r *InstansiRepository) CountAnggotaAktif(ctx context.Context, instansiID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("anggota").
		Where("instansi_id = ? AND is_deleted = false", instansiID).
		Count(&count).Error
	return count, err
}

// FileExists checks whether a mst_file row exists and is not soft-deleted.
func (r *InstansiRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
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
	"kode": true, "nama": true, "status": true, "created_at": true,
}

func (r *InstansiRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("mst_instansi.created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("mst_instansi.created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("mst_instansi.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List, bridging from the DTO.
// It is exported so the service layer can construct it.
type ListQuery struct {
	Q      string
	Status *int
	Sort   string
	Offset int
	Limit  int
}
