// Package repository provides GORM data access for the unit module.
// All queries are scoped to is_deleted = false and optionally to milik_sendiri
// (anggota_id filter based on JWT claims).
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/unit/domain"
	"slam-team-api/internal/modules/core/unit/dto"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// UnitRepository wraps *gorm.DB for unit operations.
type UnitRepository struct {
	db *gorm.DB
}

// NewUnitRepository binds a repository to the process-wide DB pool.
func NewUnitRepository(db *gorm.DB) *UnitRepository {
	return &UnitRepository{db: db}
}

// DB returns the underlying *gorm.DB for direct queries (e.g., anggota lookup).
func (r *UnitRepository) DB() *gorm.DB {
	return r.db
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *UnitRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Unit{}).Where("unit.is_deleted = false")
}

// listJoin applies LEFT JOINs for anggota name/no_induk, foto sampul UUID,
// and approver name. Caller must NOT have already joined unit again.
func (r *UnitRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			unit.*,
			a.nama_lengkap              AS anggota_nama,
			a.no_induk                  AS anggota_no_induk,
			fs.uuid::text               AS foto_sampul_uuid,
		u_approver.username          AS disetujui_oleh_nama
		`).
		Joins("LEFT JOIN anggota a ON a.id = unit.anggota_id AND a.is_deleted = false").
		Joins("LEFT JOIN mst_file fs ON fs.id = unit.foto_sampul_file_id AND fs.is_deleted = false").
		Joins("LEFT JOIN users u_approver ON u_approver.id = unit.disetujui_oleh AND u_approver.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *UnitRepository) Create(ctx context.Context, u *domain.Unit) error {
	return r.db.WithContext(ctx).Omit(
		"anggota_nama", "anggota_no_induk",
		"foto_sampul_uuid", "disetujui_oleh_nama",
	).Create(u).Error
}

// Update persists all mutable fields of an existing row.
// Uses Omit to exclude joined/virtual columns.
func (r *UnitRepository) Update(ctx context.Context, u *domain.Unit) error {
	return r.db.WithContext(ctx).Omit(
		"anggota_nama", "anggota_no_induk",
		"foto_sampul_uuid", "disetujui_oleh_nama",
	).Save(u).Error
}

// SoftDelete marks a row as deleted. The caller (service) is responsible for
// FK guard checks.
func (r *UnitRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Unit{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one unit by id, with joined anggota/foto/approver data.
func (r *UnitRepository) FindByID(ctx context.Context, id int64) (*domain.Unit, error) {
	var u domain.Unit
	q := r.baseQuery().Where("unit.id = ?", id)
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("unit: %w", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &u, nil
}

// FindByIDRaw returns one unit by id WITHOUT joins (for update snapshots).
func (r *UnitRepository) FindByIDRaw(ctx context.Context, id int64) (*domain.Unit, error) {
	var u domain.Unit
	if err := r.baseQuery().WithContext(ctx).Where("unit.id = ?", id).First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("unit: %w", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &u, nil
}

// List returns a paginated slice of units with joins.
func (r *UnitRepository) List(ctx context.Context, q ListQuery) ([]domain.Unit, int64, error) {
	base := r.baseQuery()

	// Filter by search query (ILIKE on kode, model, deskripsi_warna).
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(unit.kode) LIKE ? OR LOWER(unit.model) LIKE ? OR LOWER(unit.deskripsi_warna) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by anggota_id.
	if q.AnggotaID != nil {
		base = base.Where("unit.anggota_id = ?", *q.AnggotaID)
	}

	// Filter by disetujui.
	if q.Disetujui != nil {
		base = base.Where("unit.disetujui = ?", *q.Disetujui)
	}

	// Count total rows (before joins).
	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sort.
	base = r.applySort(base, q.Sort)

	// Paginated query with joins.
	paged := r.listJoin(base).Offset(q.Offset).Limit(q.Limit)

	var items []domain.Unit
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ── Validation helpers ──

// AnggotaExists checks if an active anggota row exists.
func (r *UnitRepository) AnggotaExists(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("anggota").
		Where("id = ? AND is_deleted = false", anggotaID).
		Count(&count).Error
	return count > 0, err
}

// FileExists checks if a non-deleted mst_file row exists by ID.
func (r *UnitRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}

// FindFileByUUID resolves a UUID string to mst_file.id. Returns 0 if not found.
func (r *UnitRepository) FindFileByUUID(ctx context.Context, uuid string) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_file").
		Select("id").
		Where("uuid = CAST(? AS uuid) AND is_deleted = false", uuid).
		Scan(&id).Error
	return id, err
}

// AnggotaTersedia returns all active anggota for dropdowns.
func (r *UnitRepository) AnggotaTersedia(ctx context.Context, items *[]dto.AnggotaOpsi) error {
	return r.db.WithContext(ctx).
		Table("anggota").
		Select("id, nama_lengkap, no_induk").
		Where("is_deleted = false").
		Order("nama_lengkap ASC").
		Scan(items).Error
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"kode": true, "model": true, "fps": true,
	"disetujui": true, "created_at": true,
}

func (r *UnitRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("unit.created_at DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("unit.created_at DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("unit.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List.
type ListQuery struct {
	Q        string
	AnggotaID *int64
	Disetujui *bool
	Sort     string
	Offset   int
	Limit    int
}
