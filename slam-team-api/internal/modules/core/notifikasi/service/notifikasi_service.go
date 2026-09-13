// Package service is the notifikasi business layer in two halves:
//
//   - the PRODUCER side (Kirim / KirimKeAnggota) is what other modules call to
//     fan a message out — penugasan today; izin, absensi and kta later. It is
//     the only sanctioned way to write a notifikasi row.
//   - the CONSUMER side (List / JumlahBelumDibaca / TandaiDibaca /
//     TandaiSemuaDibaca) serves the inbox and is always scoped to one user.
package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"slam-team-api/internal/modules/core/notifikasi/domain"
	"slam-team-api/internal/modules/core/notifikasi/dto"
	"slam-team-api/internal/modules/core/notifikasi/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/pagination"

	"gorm.io/gorm"
)

// Pesan is what a producer describes once; Kirim turns it into one row per
// recipient. Only Tipe and Judul are required.
type Pesan struct {
	// Tipe is the machine-readable kind (domain.TipeJadwalDitugaskan, …).
	Tipe string
	// Judul is the headline, max 200 chars (longer values are truncated).
	Judul string
	// Isi is the optional body text.
	Isi string
	// Ikon / Warna are UI hints; Warna is one of domain.Warna* (info default).
	Ikon  string
	Warna string
	// Route is where the client navigates on click, e.g. /jadwal/12/sesi/340.
	Route string
	// ReffType / ReffID point at the source entity (polymorphic, optional).
	ReffType string
	ReffID   *int64
	// Prioritas defaults to normal.
	Prioritas domain.Prioritas
	// KedaluwarsaPada lets the sweeper drop rows that stop being relevant
	// (a session reminder after the session is over). nil = never.
	KedaluwarsaPada *time.Time
	// CreatedBy is the triggering user; nil for system-generated rows.
	CreatedBy *int64
}

// NotifikasiService composes the repository for producer + consumer use.
type NotifikasiService struct {
	repo *repository.NotifikasiRepository
}

// NewNotifikasiService creates a service bound to the repository.
func NewNotifikasiService(repo *repository.NotifikasiRepository) *NotifikasiService {
	return &NotifikasiService{repo: repo}
}

// WithTx returns a copy whose writes run on the given transaction, so a
// producer can commit its own change and the fan-out atomically:
//
//	db.Transaction(func(tx *gorm.DB) error {
//	    … insert the business rows …
//	    _, err := notif.WithTx(tx).Kirim(ctx, userIDs, pesan)
//	    return err
//	})
func (s *NotifikasiService) WithTx(tx *gorm.DB) *NotifikasiService {
	return &NotifikasiService{repo: s.repo.WithTx(tx)}
}

// ── Producer side ──

// Kirim fans one message out to every user in penerima (duplicates collapsed)
// and returns how many rows were written. An empty recipient list is a no-op,
// not an error, so callers can pass whatever the account lookup yielded.
func (s *NotifikasiService) Kirim(ctx context.Context, penerima []int64, p Pesan) (int, error) {
	if p.Tipe == "" || p.Judul == "" {
		return 0, fmt.Errorf("notifikasi: tipe dan judul wajib diisi: %w", apperr.ErrValidation)
	}
	ids := uniqueIDs(penerima)
	if len(ids) == 0 {
		return 0, nil
	}

	rows := make([]domain.Notifikasi, 0, len(ids))
	for _, uid := range ids {
		rows = append(rows, p.toRow(uid))
	}
	if err := s.repo.CreateMany(ctx, rows); err != nil {
		return 0, fmt.Errorf("notifikasi: simpan fan-out: %w", err)
	}
	return len(rows), nil
}

// KirimKeAnggota is Kirim keyed by anggota: it resolves each anggota's live
// user account and silently drops the account-less ones (they have no inbox).
// It returns how many rows were written — which can be fewer than
// len(anggotaIDs), and that is the expected outcome, not an error.
func (s *NotifikasiService) KirimKeAnggota(ctx context.Context, anggotaIDs []int64, p Pesan) (int, error) {
	users, err := s.repo.UserIDsByAnggota(ctx, uniqueIDs(anggotaIDs))
	if err != nil {
		return 0, fmt.Errorf("notifikasi: cari akun anggota: %w", err)
	}
	penerima := make([]int64, 0, len(users))
	for _, uid := range users {
		penerima = append(penerima, uid)
	}
	return s.Kirim(ctx, penerima, p)
}

// UserIDsByAnggota exposes the account lookup for producers that need to know
// who will (and will not) receive a message before they act on it — the
// penugasan module uses it to force wajib_absen=false on account-less anggota.
func (s *NotifikasiService) UserIDsByAnggota(ctx context.Context, anggotaIDs []int64) (map[int64]int64, error) {
	return s.repo.UserIDsByAnggota(ctx, uniqueIDs(anggotaIDs))
}

// ── Consumer side (always scoped to one user) ──

// List returns one page of the user's inbox.
func (s *NotifikasiService) List(ctx context.Context, userID int64, q dto.ListNotifikasiQuery) ([]domain.Notifikasi, int64, error) {
	lq := pagination.ListQuery{Page: q.Page, PerPage: q.PerPage}
	lq.Normalize()
	items, total, err := s.repo.ListByUser(ctx, userID, repository.ListFilter{
		IsDibaca:     q.IsDibaca,
		Tipe:         q.Tipe,
		IncludeArsip: q.Arsip,
		Offset:       lq.Offset(),
		Limit:        lq.Limit(),
	})
	if err != nil {
		return nil, 0, err
	}
	if items == nil {
		items = []domain.Notifikasi{}
	}
	return items, total, nil
}

// JumlahBelumDibaca is the badge count.
func (s *NotifikasiService) JumlahBelumDibaca(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CountUnread(ctx, userID)
}

// TandaiDibaca marks one of the user's rows read and returns it. A row that
// is not the caller's is reported as not found — never as forbidden, which
// would confirm the row exists. Marking an already-read row is idempotent.
func (s *NotifikasiService) TandaiDibaca(ctx context.Context, userID, id int64) (*domain.Notifikasi, error) {
	if _, err := s.findOwn(ctx, userID, id); err != nil {
		return nil, err
	}
	if err := s.repo.MarkRead(ctx, userID, id); err != nil {
		return nil, err
	}
	return s.findOwn(ctx, userID, id)
}

// TandaiSemuaDibaca marks every unread row of the user and returns the count.
func (s *NotifikasiService) TandaiSemuaDibaca(ctx context.Context, userID int64) (int64, error) {
	return s.repo.MarkAllRead(ctx, userID)
}

func (s *NotifikasiService) findOwn(ctx context.Context, userID, id int64) (*domain.Notifikasi, error) {
	n, err := s.repo.FindByUser(ctx, userID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: notifikasi", apperr.ErrNotFound)
		}
		return nil, err
	}
	return n, nil
}

// ── Helpers ──

// toRow materialises the message for one recipient, applying column limits
// and defaults so a producer never has to think about varchar widths.
func (p Pesan) toRow(userID int64) domain.Notifikasi {
	row := domain.Notifikasi{
		UserID:    userID,
		Tipe:      truncate(p.Tipe, 50),
		Judul:     truncate(p.Judul, 200),
		Isi:       optional(p.Isi, 0),
		Ikon:      optional(p.Ikon, 50),
		Warna:     optional(p.Warna, 20),
		Route:     optional(p.Route, 255),
		ReffType:  optional(p.ReffType, 50),
		ReffID:    p.ReffID,
		Prioritas: p.Prioritas,
		CreatedBy: p.CreatedBy,
	}
	if row.Prioritas == "" {
		row.Prioritas = domain.PrioritasNormal
	}
	if row.Warna == nil {
		w := domain.WarnaInfo
		row.Warna = &w
	}
	if p.KedaluwarsaPada != nil {
		t := p.KedaluwarsaPada.UTC()
		row.KedaluwarsaPada = &t
	}
	return row
}

// optional turns an empty string into NULL and clamps to max runes (0 = no
// limit).
func optional(s string, max int) *string {
	if s == "" {
		return nil
	}
	if max > 0 {
		s = truncate(s, max)
	}
	return &s
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// uniqueIDs drops zero/negative ids and duplicates, keeping a stable order so
// the fan-out rows are inserted in a deterministic sequence.
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
