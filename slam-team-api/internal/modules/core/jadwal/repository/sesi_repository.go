package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/jadwal/domain"
	"slam-team-api/internal/shared/apperr"

	"gorm.io/gorm"
)

// sesiBatchChunk caps rows per INSERT so a multi-year daily schedule never
// approaches PostgreSQL's 65535-parameter limit (9 params per row).
const sesiBatchChunk = 500

// dateKey formats a date column value as YYYY-MM-DD — the key the service
// uses to match candidate dates against what the database reports back.
func dateKey(t time.Time) string { return t.Format("2006-01-02") }

// CreatedSesi is what CreateBatch reports for each row that was actually
// inserted; rows that hit the (jadwal_id, tanggal_lokal) unique key are
// silently skipped by ON CONFLICT DO NOTHING and therefore absent here.
type CreatedSesi struct {
	ID           int64     `gorm:"column:id"`
	TanggalLokal time.Time `gorm:"column:tanggal_lokal"`
}

// CreateBatch inserts sessions idempotently: an existing (jadwal_id,
// tanggal_lokal) pair is left untouched, so re-running generate-sesi never
// duplicates. It runs in one transaction and returns the rows it created.
func (r *JadwalRepository) CreateBatch(ctx context.Context, rows []domain.Sesi) ([]CreatedSesi, error) {
	created := make([]CreatedSesi, 0, len(rows))
	if len(rows) == 0 {
		return created, nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(rows); start += sesiBatchChunk {
			end := start + sesiBatchChunk
			if end > len(rows) {
				end = len(rows)
			}
			chunk := rows[start:end]

			values := make([]string, 0, len(chunk))
			args := make([]any, 0, len(chunk)*9)
			for _, s := range chunk {
				values = append(values, "(?, ?, ?, ?, ?, ?, ?, CAST(? AS status_sesi), ?)")
				args = append(args,
					s.JadwalID, dateKey(s.TanggalLokal),
					s.MulaiUTC, s.SelesaiUTC, s.Timezone,
					s.AbsenBukaUTC, s.AbsenTutupUTC,
					string(domain.SesiTerjadwal), s.CreatedBy,
				)
			}

			// status is bound as text and cast to the enum explicitly, so the
			// statement is valid under both the simple and extended protocols.
			sql := `
				INSERT INTO jadwal_sesi
					(jadwal_id, tanggal_lokal, mulai_utc, selesai_utc, timezone,
					 absen_buka_utc, absen_tutup_utc, status, created_by)
				VALUES ` + strings.Join(values, ",\n") + `
				ON CONFLICT (jadwal_id, tanggal_lokal) DO NOTHING
				RETURNING id, tanggal_lokal`

			var out []CreatedSesi
			if err := tx.Raw(sql, args...).Scan(&out).Error; err != nil {
				return err
			}
			created = append(created, out...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// SesiFilter narrows ListByJadwal.
type SesiFilter struct {
	Status        string
	TanggalDari   *string // YYYY-MM-DD
	TanggalSampai *string // YYYY-MM-DD
	// Dates, when non-empty, restricts to exactly these tanggal_lokal values
	// (used by generate-sesi to echo the candidate set back).
	Dates []string
}

// ListByJadwal returns the sessions of one schedule ordered by date.
func (r *JadwalRepository) ListByJadwal(ctx context.Context, jadwalID int64, f SesiFilter) ([]domain.Sesi, error) {
	q := r.db.WithContext(ctx).Model(&domain.Sesi{}).Where("jadwal_id = ?", jadwalID)
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.TanggalDari != nil {
		q = q.Where("tanggal_lokal >= ?", *f.TanggalDari)
	}
	if f.TanggalSampai != nil {
		q = q.Where("tanggal_lokal <= ?", *f.TanggalSampai)
	}
	if len(f.Dates) > 0 {
		q = q.Where("tanggal_lokal IN ?", f.Dates)
	}

	var items []domain.Sesi
	if err := q.Order("tanggal_lokal ASC, id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindSesiByID returns one session whose parent schedule is not soft-deleted,
// or ErrNotFound.
func (r *JadwalRepository) FindSesiByID(ctx context.Context, id int64) (*domain.Sesi, error) {
	var s domain.Sesi
	err := r.db.WithContext(ctx).Model(&domain.Sesi{}).
		Select("jadwal_sesi.*").
		Joins("JOIN jadwal ON jadwal.id = jadwal_sesi.jadwal_id AND jadwal.is_deleted = false").
		Where("jadwal_sesi.id = ?", id).
		First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: sesi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return &s, nil
}

// CancelSesi sets status=dibatalkan with the reason. State checks (already
// cancelled, already finished, already in the past) belong to the service.
func (r *JadwalRepository) CancelSesi(ctx context.Context, id int64, alasan string, actor *int64) error {
	return r.db.WithContext(ctx).Model(&domain.Sesi{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       string(domain.SesiDibatalkan),
			"alasan_batal": alasan,
			"modified_at":  time.Now().UTC(),
			"modified_by":  actor,
		}).Error
}
