// Package repository provides GORM data access for the profile_club module.
package repository

import (
	"context"
	"fmt"
	"strings"

	"slam-team-api/internal/modules/core/profileclub/domain"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/model"

	"gorm.io/gorm"
)

// ProfileClubRepository wraps *gorm.DB for profile_club operations.
type ProfileClubRepository struct {
	db *gorm.DB
}

// NewProfileClubRepository binds a repository to the process-wide DB pool.
func NewProfileClubRepository(db *gorm.DB) *ProfileClubRepository {
	return &ProfileClubRepository{db: db}
}

// ── Query helpers ──

// baseQuery returns a scope that filters soft-deleted rows.
func (r *ProfileClubRepository) baseQuery() *gorm.DB {
	return r.db.Model(&domain.ProfileClub{}).
		Where("profile_club.is_deleted = false")
}

// withFileJoins applies LEFT JOINs to resolve *_file_id → *_uuid for the
// three image fields (banner, logo_simple, logo_besar).
func (r *ProfileClubRepository) withFileJoins(q *gorm.DB) *gorm.DB {
	return q.
		Select(`
			profile_club.*,
			fb.uuid   AS banner_uuid,
			fs.uuid   AS logo_simple_uuid,
			fl.uuid   AS logo_besar_uuid
		`).
		Joins("LEFT JOIN mst_file fb ON fb.id = profile_club.banner_file_id AND fb.is_deleted = false").
		Joins("LEFT JOIN mst_file fs ON fs.id = profile_club.logo_simple_file_id AND fs.is_deleted = false").
		Joins("LEFT JOIN mst_file fl ON fl.id = profile_club.logo_besar_file_id AND fl.is_deleted = false")
}

// ── CRUD ──

// Create inserts one row and returns the populated entity.
func (r *ProfileClubRepository) Create(ctx context.Context, d *domain.ProfileClub) error {
	return r.db.WithContext(ctx).Omit("BannerUUID", "LogoSimpleUUID", "LogoBesarUUID").Create(d).Error
}

// FindByID returns one profile_club by id, with file UUIDs resolved via JOINs.
func (r *ProfileClubRepository) FindByID(ctx context.Context, id int64) (*domain.ProfileClub, error) {
	var d domain.ProfileClub
	if err := r.withFileJoins(r.baseQuery()).WithContext(ctx).
		Where("profile_club.id = ?", id).
		First(&d).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: profile club", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &d, nil
}

// GetActive returns the first active row (singleton semantics).
func (r *ProfileClubRepository) GetActive(ctx context.Context) (*domain.ProfileClub, error) {
	var d domain.ProfileClub
	err := r.withFileJoins(r.baseQuery()).WithContext(ctx).
		Order("created_at ASC").
		First(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w: profile club", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &d, nil
}

// CountActive returns the number of non-deleted rows (singleton guard).
func (r *ProfileClubRepository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	if err := r.baseQuery().WithContext(ctx).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// List returns a paginated slice (practical: 0/1 rows).
func (r *ProfileClubRepository) List(ctx context.Context, q ListQuery) ([]domain.ProfileClub, int64, error) {
	base := r.baseQuery()

	if q.Q != "" {
		like := "%" + strings.ToLower(q.Q) + "%"
		base = base.Where(
			"(LOWER(nama) LIKE ? OR LOWER(singkatan) LIKE ?)",
			like, like,
		)
	}

	var total int64
	if err := base.WithContext(ctx).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	col, desc := "created_at", true
	if q.Sort != "" {
		sortCol, sortDesc, ok := r.allowedSort(q.Sort)
		if ok {
			col, desc = sortCol, sortDesc
		}
	}
	dir := "DESC"
	if !desc {
		dir = "ASC"
	}

	var items []domain.ProfileClub
	if err := r.withFileJoins(base).WithContext(ctx).
		Order(fmt.Sprintf("%s %s", col, dir)).
		Offset(q.Offset).Limit(q.Limit).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// Update persists all mutable fields of an existing row.
func (r *ProfileClubRepository) Update(ctx context.Context, d *domain.ProfileClub) error {
	return r.db.WithContext(ctx).Omit("BannerUUID", "LogoSimpleUUID", "LogoBesarUUID").Save(d).Error
}

// SoftDelete marks a row as deleted.
func (r *ProfileClubRepository) SoftDelete(ctx context.Context, id int64, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.ProfileClub{}).
		Where("id = ? AND is_deleted = false", id).
		Updates(model.SoftDeleteFields(actor)).Error
}

// ── Sort ──

var allowedSortCols = map[string]bool{
	"created_at": true, "modified_at": true, "nama": true,
}

func (r *ProfileClubRepository) allowedSort(sort string) (string, bool, bool) {
	col, desc := sort, false
	if col[0] == '-' {
		col, desc = col[1:], true
	}
	if allowedSortCols[col] {
		return col, desc, true
	}
	return "", false, false
}

// ── Public prestasi aggregation ──

// PublicPrestasi is the lightweight prestasi shape for the public endpoint.
type PublicPrestasi struct {
	JudulKompetisi   string  `json:"judul_kompetisi"`
	Peringkat        *string `json:"peringkat,omitempty"`
	Tingkat          *string `json:"tingkat,omitempty"`
	TanggalKompetisi *string `json:"tanggal_kompetisi,omitempty"`
}

// GetPublicPrestasi returns Guest-readable prestasi for the public profile.
func (r *ProfileClubRepository) GetPublicPrestasi(ctx context.Context) ([]PublicPrestasi, error) {
	var items []PublicPrestasi
	err := r.db.WithContext(ctx).
		Table("prestasi").
		Select(`
			prestasi.judul_kompetisi,
			prestasi.peringkat,
			prestasi.tingkat,
			TO_CHAR(prestasi.tanggal_kompetisi, 'YYYY-MM-DD') AS tanggal_kompetisi
		`).
		Where("prestasi.is_deleted = false").
		Order("prestasi.tanggal_kompetisi DESC").
		Limit(50).
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}

// ListQuery is the internal query shape used by List.
type ListQuery struct {
	Q       string
	Sort    string
	Offset  int
	Limit   int
}

// FileExists checks whether a mst_file row exists and is not soft-deleted.
func (r *ProfileClubRepository) FileExists(ctx context.Context, fileID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&count).Error
	return count > 0, err
}
