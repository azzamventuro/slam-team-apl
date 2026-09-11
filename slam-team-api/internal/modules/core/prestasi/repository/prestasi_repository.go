// Package repository provides GORM data access for the prestasi module.
// All queries are scoped to is_deleted = false.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/prestasi/domain"
	"slam-team-api/internal/modules/core/prestasi/dto"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// PrestasiRepository wraps *gorm.DB for prestasi operations.
type PrestasiRepository struct {
	db *gorm.DB
}

// NewPrestasiRepository binds a repository to the process-wide DB pool.
func NewPrestasiRepository(db *gorm.DB) *PrestasiRepository {
	return &PrestasiRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *PrestasiRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Prestasi{}).Where("prestasi.is_deleted = false")
}

// listJoin applies LEFT JOINs for anggota name/no_induk, flyer UUID, foto sampul UUID.
func (r *PrestasiRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			prestasi.*,
			a.nama_lengkap              AS anggota_nama,
			a.no_induk                  AS anggota_no_induk,
			COALESCE(a.nama_panggilan, '') AS anggota_nama_panggilan,
			ff.id                       AS flyer_file_id_ref,
			ff.uuid::text               AS flyer_uuid,
			ff.is_publik                AS flyer_is_publik,
			fs.uuid::text               AS foto_sampul_uuid,
			fs.is_publik                AS foto_sampul_is_publik
		`).
		Joins("LEFT JOIN anggota a ON a.id = prestasi.anggota_id AND a.is_deleted = false").
		Joins("LEFT JOIN mst_file ff ON ff.id = prestasi.flyer_file_id AND ff.is_deleted = false").
		Joins("LEFT JOIN mst_file fs ON fs.id = prestasi.foto_sampul_file_id AND fs.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row.
func (r *PrestasiRepository) Create(ctx context.Context, p *domain.Prestasi) error {
	return r.db.WithContext(ctx).Omit(
		"anggota_nama", "anggota_no_induk", "anggota_nama_panggilan",
		"flyer_file_id_ref", "flyer_uuid", "flyer_is_publik",
		"foto_sampul_uuid", "foto_sampul_is_publik",
	).Create(p).Error
}

// Update persists all mutable fields of an existing row.
func (r *PrestasiRepository) Update(ctx context.Context, p *domain.Prestasi) error {
	return r.db.WithContext(ctx).Omit(
		"anggota_nama", "anggota_no_induk", "anggota_nama_panggilan",
		"flyer_file_id_ref", "flyer_uuid", "flyer_is_publik",
		"foto_sampul_uuid", "foto_sampul_is_publik",
	).Save(p).Error
}

// SoftDelete marks a row as deleted.
func (r *PrestasiRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Prestasi{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one prestasi by id, with joined data.
func (r *PrestasiRepository) FindByID(ctx context.Context, id int64) (*domain.Prestasi, error) {
	var p domain.Prestasi
	q := r.baseQuery().Where("prestasi.id = ?", id)
	q = r.listJoin(q)
	if err := q.WithContext(ctx).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("prestasi: %w", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &p, nil
}

// List returns a paginated slice of prestasi with joins.
func (r *PrestasiRepository) List(ctx context.Context, q ListQuery) ([]domain.Prestasi, int64, error) {
	base := r.baseQuery()

	// Filter by search query (ILIKE on judul_kompetisi, tingkat, peringkat).
	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(prestasi.judul_kompetisi) LIKE ? OR LOWER(prestasi.tingkat) LIKE ? OR LOWER(prestasi.peringkat) LIKE ?)",
			like, like, like,
		)
	}

	// Filter by anggota_id.
	if q.AnggotaID != nil {
		base = base.Where("prestasi.anggota_id = ?", *q.AnggotaID)
	}

	// Filter by tingkat.
	if q.Tingkat != nil && *q.Tingkat != "" {
		base = base.Where("LOWER(prestasi.tingkat) = LOWER(?)", *q.Tingkat)
	}

	// Filter by peringkat.
	if q.Peringkat != nil && *q.Peringkat != "" {
		base = base.Where("LOWER(prestasi.peringkat) = LOWER(?)", *q.Peringkat)
	}

	// Filter by tanggal range.
	if q.TanggalDari != nil && *q.TanggalDari != "" {
		base = base.Where("prestasi.tanggal_kompetisi >= ?", *q.TanggalDari)
	}
	if q.TanggalSampai != nil && *q.TanggalSampai != "" {
		base = base.Where("prestasi.tanggal_kompetisi <= ?", *q.TanggalSampai)
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

	var items []domain.Prestasi
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// PublicList returns paginated prestasi safe for public consumption.
func (r *PrestasiRepository) PublicList(ctx context.Context, q ListQuery) ([]domain.Prestasi, int64, error) {
	// Reuse List — public list has the same joins.
	return r.List(ctx, q)
}

// ── Validation helpers ──

// AnggotaExists checks if an active anggota row exists.
func (r *PrestasiRepository) AnggotaExists(ctx context.Context, anggotaID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("anggota").
		Where("id = ? AND is_deleted = false", anggotaID).
		Count(&count).Error
	return count > 0, err
}

// FileExists checks if a non-deleted mst_file row exists by ID.
func (r *PrestasiRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&struct{}{}).
		Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}

// AnggotaTersedia returns all active anggota for dropdowns.
func (r *PrestasiRepository) AnggotaTersedia(ctx context.Context, items *[]dto.AnggotaOpsi) error {
	return r.db.WithContext(ctx).
		Table("anggota").
		Select("id, nama_lengkap, no_induk").
		Where("is_deleted = false").
		Order("nama_lengkap ASC").
		Scan(items).Error
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"tanggal_kompetisi": true,
	"created_at":        true,
	"judul_kompetisi":   true,
	"peringkat":         true,
	"tingkat":           true,
}

func (r *PrestasiRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("prestasi.tanggal_kompetisi DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("prestasi.tanggal_kompetisi DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("prestasi.%s %s", col, dir))
}

// ListQuery is the internal query shape used by List.
type ListQuery struct {
	Q            string
	AnggotaID    *int64
	Tingkat      *string
	Peringkat    *string
	TanggalDari  *string
	TanggalSampai *string
	Sort         string
	Offset       int
	Limit        int
}
