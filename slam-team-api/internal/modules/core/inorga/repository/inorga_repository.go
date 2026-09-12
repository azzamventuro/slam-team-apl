package repository

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"slam-team-api/internal/modules/core/inorga/domain"
	"slam-team-api/internal/shared/apperr"
)

// InorgaRepository encapsulates GORM access for mst_inorga.
type InorgaRepository struct{ db *gorm.DB }

func NewInorgaRepository(db *gorm.DB) *InorgaRepository {
	return &InorgaRepository{db: db}
}

// ── helpers ──

func (r *InorgaRepository) baseQuery() *gorm.DB {
	return r.db.Table("mst_inorga").Where("mst_inorga.is_deleted = false")
}

// listJoin adds LEFT JOINs for file uuids + nama_asli.
func (r *InorgaRepository) listJoin(q *gorm.DB) *gorm.DB {
	return q.
		Joins(`LEFT JOIN mst_file lf ON lf.id = mst_inorga.logo_file_id AND lf.is_deleted = false`).
		Joins(`LEFT JOIN mst_file bf ON bf.id = mst_inorga.banner_file_id AND bf.is_deleted = false`).
		Joins(`LEFT JOIN mst_file sf ON sf.id = mst_inorga.file_sk_file_id AND sf.is_deleted = false`)
}

// selectCols returns the full SELECT expression including computed `aktif`.
func (r *InorgaRepository) selectCols() string {
	return `
		mst_inorga.*,
		lf.uuid::text       AS logo_uuid,
		lf.nama_asli        AS logo_nama_asli,
		bf.uuid::text       AS banner_uuid,
		bf.nama_asli        AS banner_nama_asli,
		sf.uuid::text       AS file_sk_uuid,
		sf.nama_asli        AS file_sk_nama_asli,
		CASE
			WHEN mst_inorga.tanggal_mulai IS NOT NULL
				AND mst_inorga.tanggal_mulai <= CURRENT_DATE
				AND (mst_inorga.tanggal_selesai IS NULL OR mst_inorga.tanggal_selesai >= CURRENT_DATE)
			THEN true ELSE false
		END AS aktif`
}

// ── CRUD ──

func (r *InorgaRepository) Create(i *domain.Inorga) error {
	return r.db.Table("mst_inorga").Omit(
		"logo_uuid", "banner_uuid", "file_sk_uuid",
	).Create(i).Error
}

func (r *InorgaRepository) FindByID(id int64) (*domain.Inorga, error) {
	var out domain.Inorga
	err := r.listJoin(r.baseQuery()).
		Select(r.selectCols()).
		Where("mst_inorga.id = ?", id).
		First(&out).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("inorga: %w", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &out, nil
}

type ListQuery struct {
	Page    int
	PerPage int
	Q       string
	Sort    string
	Aktif   string
}

// ListQuery returns a page of inorga records.
func (r *InorgaRepository) List(lq ListQuery) ([]domain.Inorga, int64, error) {
	q := r.listJoin(r.baseQuery()).Select(r.selectCols())

	// ── filters ──
	if lq.Q != "" {
		like := "%" + lq.Q + "%"
		q = q.Where("(mst_inorga.nama ILIKE ? OR mst_inorga.kode ILIKE ?)", like, like)
	}
	if lq.Aktif == "true" {
		q = q.Where(`mst_inorga.tanggal_mulai IS NOT NULL
			AND mst_inorga.tanggal_mulai <= CURRENT_DATE
			AND (mst_inorga.tanggal_selesai IS NULL OR mst_inorga.tanggal_selesai >= CURRENT_DATE)`)
	} else if lq.Aktif == "false" {
		q = q.Where(`mst_inorga.tanggal_mulai IS NULL
			OR mst_inorga.tanggal_mulai > CURRENT_DATE
			OR (mst_inorga.tanggal_selesai IS NOT NULL AND mst_inorga.tanggal_selesai < CURRENT_DATE)`)
	}

	// ── count (before pagination) ──
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// ── sort ──
	q = q.Order(applySort(lq.Sort))

	// ── paginate ──
	offset := (lq.Page - 1) * lq.PerPage
	var rows []domain.Inorga
	if err := q.Offset(offset).Limit(lq.PerPage).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *InorgaRepository) Update(i *domain.Inorga) error {
	return r.db.Table("mst_inorga").
		Omit("id", "created_at", "created_by", "logo_uuid", "banner_uuid", "file_sk_uuid").
		Save(i).Error
}

func (r *InorgaRepository) SoftDelete(id int64, deletedBy int64) error {
	now := time.Now()
	return r.db.Table("mst_inorga").
		Where("id = ? AND is_deleted = false", id).
		Updates(map[string]any{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": deletedBy,
		}).Error
}

// ── Validators ──

func (r *InorgaRepository) FileExists(fileID int64) (bool, error) {
	var cnt int64
	err := r.db.Table("mst_file").
		Where("id = ? AND is_deleted = false", fileID).
		Count(&cnt).Error
	return cnt > 0, err
}

// ── sort ──

var inorgaSortMap = map[string]string{
	"created_at":   "mst_inorga.created_at",
	"tanggal_mulai": "mst_inorga.tanggal_mulai",
	"nama":          "mst_inorga.nama",
}

func applySort(raw string) string {
	col := "mst_inorga.created_at"
	desc := true
	if raw != "" {
		asc := raw[0] != '-'
		key := raw
		if !asc {
			key = raw[1:]
		}
		if mapped, ok := inorgaSortMap[key]; ok {
			col = mapped
			desc = !asc
		}
	}
	dir := "DESC"
	if !desc {
		dir = "ASC"
	}
	return fmt.Sprintf("%s %s", col, dir)
}
