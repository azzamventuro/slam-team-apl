// Package repository provides GORM data access for the jadwal module:
// the jadwal table (schedule definitions) and jadwal_sesi (materialised
// sessions), plus the cross-table existence checks the service needs.
package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/jadwal/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// JadwalRepository wraps *gorm.DB for jadwal + jadwal_sesi operations.
type JadwalRepository struct {
	db *gorm.DB
}

// NewJadwalRepository binds a repository to the process-wide DB pool.
func NewJadwalRepository(db *gorm.DB) *JadwalRepository {
	return &JadwalRepository{db: db}
}

// ── Query helpers ──

// baseQuery filters soft-deleted rows. is_deleted is table-qualified because
// the detail/list reads LEFT JOIN mst_lokasi, which carries the same column.
func (r *JadwalRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.Jadwal{}).Where("jadwal.is_deleted = false")
}

// listJoin resolves the read-only projections (lokasi_nama, jumlah_sesi).
// Call before Find/First, never before Count.
func (r *JadwalRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			jadwal.*,
			l.nama AS lokasi_nama,
			(SELECT count(*) FROM jadwal_sesi s WHERE s.jadwal_id = jadwal.id) AS jumlah_sesi
		`).
		Joins("LEFT JOIN mst_lokasi l ON l.id = jadwal.lokasi_id AND l.is_deleted = false")
}

// ── Jadwal CRUD ──

// Create inserts one schedule and populates its ID.
func (r *JadwalRepository) Create(ctx context.Context, j *domain.Jadwal) error {
	return r.db.WithContext(ctx).Omit("lokasi_nama", "jumlah_sesi").Create(j).Error
}

// Update persists every mutable column of an existing schedule.
func (r *JadwalRepository) Update(ctx context.Context, j *domain.Jadwal) error {
	return r.db.WithContext(ctx).Omit("lokasi_nama", "jumlah_sesi").Save(j).Error
}

// SoftDelete marks a schedule deleted. Guards belong to the service.
func (r *JadwalRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Jadwal{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// FindByID returns one schedule with its projections, or ErrNotFound.
func (r *JadwalRepository) FindByID(ctx context.Context, id int64) (*domain.Jadwal, error) {
	var j domain.Jadwal
	q := r.listJoin(r.baseQuery().Where("jadwal.id = ?", id))
	if err := q.WithContext(ctx).First(&j).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: jadwal", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &j, nil
}

// ListQuery is the internal filter shape for List, built by the service.
type ListQuery struct {
	Q             string
	Status        string
	LokasiID      *int64
	PolaUlang     string
	TanggalDari   *string // YYYY-MM-DD
	TanggalSampai *string // YYYY-MM-DD
	Sort          string
	Offset        int
	Limit         int
}

// List returns one page of schedules plus the unpaged total.
func (r *JadwalRepository) List(ctx context.Context, q ListQuery) ([]domain.Jadwal, int64, error) {
	base := r.baseQuery()

	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(jadwal.kode) LIKE ? OR LOWER(jadwal.nama) LIKE ? OR LOWER(COALESCE(jadwal.jenis_jadwal, '')) LIKE ?)",
			like, like, like,
		)
	}
	if q.Status != "" {
		base = base.Where("jadwal.status = ?", q.Status)
	}
	if q.LokasiID != nil {
		base = base.Where("jadwal.lokasi_id = ?", *q.LokasiID)
	}
	if q.PolaUlang != "" {
		base = base.Where("jadwal.pola_ulang = ?", q.PolaUlang)
	}
	// Overlap test: the schedule's active span intersects [dari, sampai].
	if q.TanggalSampai != nil {
		base = base.Where("jadwal.tanggal_mulai <= ?", *q.TanggalSampai)
	}
	if q.TanggalDari != nil {
		base = base.Where(
			"COALESCE(jadwal.tanggal_akhir_ulang, jadwal.tanggal_selesai, jadwal.tanggal_mulai) >= ?",
			*q.TanggalDari,
		)
	}

	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	base = r.applySort(base, q.Sort)
	paged := r.listJoin(base).Offset(q.Offset).Limit(q.Limit)

	var items []domain.Jadwal
	if err := paged.WithContext(ctx).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// ExistsKode reports whether kode is taken by another non-deleted schedule.
func (r *JadwalRepository) ExistsKode(ctx context.Context, kode string, exceptID int64) (bool, error) {
	var count int64
	q := r.baseQuery().Where("jadwal.kode = ?", kode)
	if exceptID > 0 {
		q = q.Where("jadwal.id <> ?", exceptID)
	}
	if err := q.WithContext(ctx).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ── Sort ──

var allowedSortColumns = map[string]bool{
	"kode": true, "nama": true, "tanggal_mulai": true, "status": true, "created_at": true,
}

func (r *JadwalRepository) applySort(q *gorm.DB, sort string) *gorm.DB {
	if sort == "" {
		return q.Order("jadwal.tanggal_mulai DESC, jadwal.id DESC")
	}
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if !allowedSortColumns[col] {
		return q.Order("jadwal.tanggal_mulai DESC, jadwal.id DESC")
	}
	dir := "ASC"
	if desc {
		dir = "DESC"
	}
	return q.Order(fmt.Sprintf("jadwal.%s %s, jadwal.id %s", col, dir, dir))
}

// ── Foreign-row lookups ──

// LokasiSnapshot is the slice of mst_lokasi a schedule copies at creation.
type LokasiSnapshot struct {
	ID          int64   `gorm:"column:id"`
	Latitude    float64 `gorm:"column:latitude"`
	Longitude   float64 `gorm:"column:longitude"`
	RadiusMeter int     `gorm:"column:radius_meter"`
	Timezone    string  `gorm:"column:timezone"`
}

// FindLokasi returns the geofence snapshot of a non-deleted mst_lokasi, or
// ErrNotFound.
func (r *JadwalRepository) FindLokasi(ctx context.Context, id int64) (*LokasiSnapshot, error) {
	var l LokasiSnapshot
	err := r.db.WithContext(ctx).
		Table("mst_lokasi").
		Select("id, latitude, longitude, radius_meter, timezone").
		Where("id = ? AND is_deleted = false", id).
		Take(&l).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: lokasi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &l, nil
}

// InorgaExists reports whether a non-deleted mst_inorga row exists.
func (r *JadwalRepository) InorgaExists(ctx context.Context, id int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("mst_inorga").
		Where("id = ? AND is_deleted = false", id).
		Count(&count).Error
	return count > 0, err
}
