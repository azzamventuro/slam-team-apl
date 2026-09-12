// Package repository provides GORM data access for the anggota module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/anggota/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// AnggotaRepository wraps *gorm.DB for anggota operations.
type AnggotaRepository struct {
	db *gorm.DB
}

// NewAnggotaRepository binds a repository to the process-wide DB pool.
func NewAnggotaRepository(db *gorm.DB) *AnggotaRepository {
	return &AnggotaRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *AnggotaRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Anggota{}).Where("anggota.is_deleted = false")
}

// listJoin applies LEFT JOINs for instansi nama, wilayah nama, and file UUIDs.
func (r *AnggotaRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			anggota.*,
			mi.nama       AS instansi_nama,
			mw.nama       AS wilayah_nama,
			fp.uuid       AS foto_profil_uuid,
			ff.uuid       AS foto_formal_uuid,
			fi.uuid       AS file_identitas_uuid
		`).
		Joins("LEFT JOIN mst_instansi mi ON mi.id = anggota.instansi_id AND mi.is_deleted = false").
		Joins("LEFT JOIN mst_wilayah mw ON mw.id = anggota.wilayah_id").
		Joins("LEFT JOIN mst_file fp ON fp.id = anggota.foto_profil_file_id AND fp.is_deleted = false").
		Joins("LEFT JOIN mst_file ff ON ff.id = anggota.foto_formal_file_id AND ff.is_deleted = false").
		Joins("LEFT JOIN mst_file fi ON fi.id = anggota.file_identitas_file_id AND fi.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *AnggotaRepository) Create(ctx context.Context, a *domain.Anggota) error {
	return r.db.WithContext(ctx).Create(a).Error
}

// Update persists all mutable fields of an existing row.
func (r *AnggotaRepository) Update(ctx context.Context, a *domain.Anggota) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// SoftDelete marks a row as deleted.
func (r *AnggotaRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Anggota{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one anggota by id with joined fields.
func (r *AnggotaRepository) FindByID(ctx context.Context, id int64) (*domain.Anggota, error) {
	var a domain.Anggota
	q := r.baseQuery().Where("anggota.id = ?", id)
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: anggota", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &a, nil
}

// List returns a paginated slice of anggota with joined fields.
func (r *AnggotaRepository) List(ctx context.Context, q ListQuery) ([]domain.Anggota, int64, error) {
	base := r.baseQuery()

	// Filter by search query (ILIKE on nama_lengkap, no_induk, nama_panggilan).
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(anggota.nama_lengkap) LIKE ? OR LOWER(anggota.no_induk) LIKE ? OR LOWER(anggota.nama_panggilan) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by instansi_id.
	if q.InstansiID != nil {
		base = base.Where("anggota.instansi_id = ?", *q.InstansiID)
	}

	// Filter by jenis_anggota.
	if q.Jenis != nil {
		base = base.Where("anggota.jenis_anggota = ?", *q.Jenis)
	}

	// Filter by status_anggota.
	if q.Status != nil {
		base = base.Where("anggota.status_anggota = ?", *q.Status)
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

	var items []domain.Anggota
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ExistsNoInduk checks whether a no_induk already exists among non-deleted rows,
// optionally excluding a given id.
func (r *AnggotaRepository) ExistsNoInduk(ctx context.Context, noInduk string, exceptID int64) (bool, error) {
	var count int64
	q := r.baseQuery().Where("no_induk = ?", noInduk)
	if exceptID > 0 {
		q = q.Where("id <> ?", exceptID)
	}
	if err := q.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// InstansiExists checks whether a non-deleted instansi row exists.
func (r *AnggotaRepository) InstansiExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_instansi").
		Where("id = ? AND is_deleted = false", id).
		Count(&count).Error
	return count > 0, err
}

// WilayahExists checks whether a wilayah row exists.
func (r *AnggotaRepository) WilayahExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_wilayah").
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}

// FileExists checks whether a mst_file row exists and is not soft-deleted.
func (r *AnggotaRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}

// HasActiveKta checks whether there are active KTA rows referencing this anggota.
func (r *AnggotaRepository) HasActiveKta(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("kta").
		Where("anggota_id = ? AND is_deleted = false AND status = 'aktif'", anggotaID).
		Count(&count).Error
	return count > 0, err
}

// HasUserAccount checks whether there are active user accounts referencing this anggota.
func (r *AnggotaRepository) HasUserAccount(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Where("anggota_id = ? AND is_deleted = false", anggotaID).
		Count(&count).Error
	return count > 0, err
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"nama_lengkap": true, "no_induk": true, "created_at": true, "tanggal_bergabung": true,
}

func (r *AnggotaRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("anggota.created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("anggota.created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("anggota.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List.
type ListQuery struct {
	Q          string
	InstansiID *int64
	Jenis      *string
	Status     *string
	Sort       string
	Offset     int
	Limit      int
}
