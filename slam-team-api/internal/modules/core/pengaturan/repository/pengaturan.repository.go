// Package repository is the GORM data access for mst_pengaturan. It returns
// domain entities and never touches *gin.Context.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/pengaturan/domain"
	"slam-team-api/internal/shared/apperr"

	"gorm.io/gorm"
)

// PengaturanRepository reads and writes mst_pengaturan.
type PengaturanRepository struct {
	db *gorm.DB
}

// NewPengaturanRepository binds the repository to the pool.
func NewPengaturanRepository(db *gorm.DB) *PengaturanRepository {
	return &PengaturanRepository{db: db}
}

// ordered is the canonical read order — the settings page renders tabs by grup
// and controls by urutan within a tab.
func ordered(db *gorm.DB) *gorm.DB { return db.Order("grup ASC, urutan ASC, kunci ASC") }

// List returns every setting. There is no soft-delete scope here: the table has
// no is_deleted column (settings are edited, never removed).
func (r *PengaturanRepository) List(ctx context.Context) ([]domain.Pengaturan, error) {
	var rows []domain.Pengaturan
	err := ordered(r.db.WithContext(ctx)).Find(&rows).Error
	return rows, err
}

// ListByGrup returns the settings of one grup, empty when the grup is unknown.
func (r *PengaturanRepository) ListByGrup(ctx context.Context, grup string) ([]domain.Pengaturan, error) {
	var rows []domain.Pengaturan
	err := ordered(r.db.WithContext(ctx)).Where("grup = ?", grup).Find(&rows).Error
	return rows, err
}

// ListPublik returns only the rows flagged is_publik. The filter is in the
// query, not in the caller: GET /public/pengaturan is unauthenticated, so a
// forgotten filter upstream would expose thresholds to anyone.
func (r *PengaturanRepository) ListPublik(ctx context.Context) ([]domain.Pengaturan, error) {
	var rows []domain.Pengaturan
	err := ordered(r.db.WithContext(ctx)).Where("is_publik = true").Find(&rows).Error
	return rows, err
}

// FindByKunci looks a setting up by its natural key, translating GORM's
// not-found into the apperr sentinel.
func (r *PengaturanRepository) FindByKunci(ctx context.Context, kunci string) (domain.Pengaturan, error) {
	var row domain.Pengaturan
	err := r.db.WithContext(ctx).Where("kunci = ?", kunci).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Pengaturan{}, fmt.Errorf("pengaturan %q: %w", kunci, apperr.ErrNotFound)
	}
	return row, err
}

// BulkUpsert writes nilai for several settings in ONE transaction: either every
// key lands or none does, so a half-applied settings page is impossible.
//
// "Upsert" is the endpoint's vocabulary, not an INSERT: rows are created by the
// seeder and addressed here by their natural key. A kunci with no row is a
// programming error at this layer — the service rejects unknown keys with a 422
// before calling — so a zero-rows update aborts the transaction.
func (r *PengaturanRepository) BulkUpsert(ctx context.Context, nilai map[string]string, by *int64) error {
	if len(nilai) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for kunci, v := range nilai {
			res := tx.Model(&domain.Pengaturan{}).
				Where("kunci = ?", kunci).
				Updates(map[string]any{
					"nilai":       v,
					"modified_at": now,
					"modified_by": by,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("pengaturan %q: %w", kunci, apperr.ErrNotFound)
			}
		}
		return nil
	})
}
