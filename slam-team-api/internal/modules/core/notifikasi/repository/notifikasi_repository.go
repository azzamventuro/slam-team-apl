// Package repository is the GORM data access for notifikasi. Every read and
// every state change is keyed by user_id: a row is only ever reachable by its
// recipient.
package repository

import (
	"context"
	"time"

	"slam-team-api/internal/modules/core/notifikasi/domain"

	"gorm.io/gorm"
)

// notifBatchChunk caps rows per INSERT (18 columns per row) well under
// PostgreSQL's 65535-parameter limit.
const notifBatchChunk = 500

// NotifikasiRepository is the data-access layer for the notifikasi table.
type NotifikasiRepository struct {
	db *gorm.DB
}

// NewNotifikasiRepository creates a repository bound to the pool.
func NewNotifikasiRepository(db *gorm.DB) *NotifikasiRepository {
	return &NotifikasiRepository{db: db}
}

// WithTx returns a copy bound to a transaction handle, so a producer module
// can fan out inside the same transaction as the change that triggered it.
func (r *NotifikasiRepository) WithTx(tx *gorm.DB) *NotifikasiRepository {
	return &NotifikasiRepository{db: tx}
}

// CreateMany inserts the fan-out rows in chunks and populates their IDs/UUIDs.
func (r *NotifikasiRepository) CreateMany(ctx context.Context, rows []domain.Notifikasi) error {
	if len(rows) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(rows, notifBatchChunk).Error
}

// ListFilter narrows ListByUser.
type ListFilter struct {
	// IsDibaca, when non-nil, restricts to read (true) or unread (false) rows.
	IsDibaca *bool
	// Tipe, when non-empty, restricts to one tipe value.
	Tipe string
	// IncludeArsip includes archived rows (excluded by default).
	IncludeArsip bool
	Offset       int
	Limit        int
}

// ListByUser returns one page of a user's inbox, newest first, plus the
// unpaged total for the same filter.
func (r *NotifikasiRepository) ListByUser(ctx context.Context, userID int64, f ListFilter) ([]domain.Notifikasi, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.Notifikasi{}).Where("user_id = ?", userID)
	if f.IsDibaca != nil {
		q = q.Where("is_dibaca = ?", *f.IsDibaca)
	}
	if f.Tipe != "" {
		q = q.Where("tipe = ?", f.Tipe)
	}
	if !f.IncludeArsip {
		q = q.Where("is_diarsipkan = false")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.Notifikasi
	err := q.Order("created_at DESC, id DESC").Offset(f.Offset).Limit(f.Limit).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// CountUnread returns the badge number: unread, non-archived rows of a user.
func (r *NotifikasiRepository) CountUnread(ctx context.Context, userID int64) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.Notifikasi{}).
		Where("user_id = ? AND is_dibaca = false AND is_diarsipkan = false", userID).
		Count(&n).Error
	return n, err
}

// FindByUser returns one row of a user's inbox, or gorm.ErrRecordNotFound —
// a row that belongs to someone else is indistinguishable from a missing one.
func (r *NotifikasiRepository) FindByUser(ctx context.Context, userID, id int64) (*domain.Notifikasi, error) {
	var n domain.Notifikasi
	err := r.db.WithContext(ctx).Model(&domain.Notifikasi{}).
		Where("id = ? AND user_id = ?", id, userID).
		First(&n).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// MarkRead flips one row to read. Keyed by (id, user_id) like every other
// access; already-read rows are left untouched (dibaca_pada keeps the first
// read time).
func (r *NotifikasiRepository) MarkRead(ctx context.Context, userID, id int64) error {
	return r.db.WithContext(ctx).Model(&domain.Notifikasi{}).
		Where("id = ? AND user_id = ? AND is_dibaca = false", id, userID).
		Updates(map[string]any{"is_dibaca": true, "dibaca_pada": time.Now().UTC()}).Error
}

// MarkAllRead flips every unread row of a user and returns how many changed.
func (r *NotifikasiRepository) MarkAllRead(ctx context.Context, userID int64) (int64, error) {
	res := r.db.WithContext(ctx).Model(&domain.Notifikasi{}).
		Where("user_id = ? AND is_dibaca = false", userID).
		Updates(map[string]any{"is_dibaca": true, "dibaca_pada": time.Now().UTC()})
	return res.RowsAffected, res.Error
}

// UserIDsByAnggota resolves the account behind each anggota. Only anggota
// with a live (non-deleted) users row appear in the result, because
// account-less members receive no in-app notification. Should an anggota
// somehow have several accounts, the oldest one wins.
func (r *NotifikasiRepository) UserIDsByAnggota(ctx context.Context, anggotaIDs []int64) (map[int64]int64, error) {
	out := map[int64]int64{}
	if len(anggotaIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		AnggotaID int64 `gorm:"column:anggota_id"`
		UserID    int64 `gorm:"column:id"`
	}
	err := r.db.WithContext(ctx).Table("users").
		Select("anggota_id, id").
		Where("anggota_id IN ? AND is_deleted = false", anggotaIDs).
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if _, dup := out[row.AnggotaID]; !dup {
			out[row.AnggotaID] = row.UserID
		}
	}
	return out, nil
}
