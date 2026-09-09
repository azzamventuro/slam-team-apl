// Package service holds the pengaturan business rules: type-checking a value
// against its tipe_nilai, the super-admin-only rule for locked rows, and the
// audit entry every successful change leaves behind.
package service

import (
	"context"
	"encoding/json"
	"fmt"

	"slam-team-api/internal/modules/core/pengaturan/domain"
	"slam-team-api/internal/modules/core/pengaturan/dto"
	"slam-team-api/internal/modules/core/pengaturan/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor is who is making the change, plus the request fingerprint the audit row
// records. UserID is nil for system/CLI actions.
type Actor struct {
	UserID    *int64
	IsSuper   bool
	IPAddress string
	UserAgent string
}

// PengaturanService reads settings and applies bulk updates.
type PengaturanService struct {
	repo  *repository.PengaturanRepository
	audit *audit.Writer
}

// NewPengaturanService composes the service. audit may be nil (tests); the
// writer tolerates it.
func NewPengaturanService(repo *repository.PengaturanRepository, w *audit.Writer) *PengaturanService {
	return &PengaturanService{repo: repo, audit: w}
}

// List returns every setting, already grouped by grup in urutan order, so the
// settings page can render one tab per group without regrouping.
func (s *PengaturanService) List(ctx context.Context) ([]dto.SettingGroup, error) {
	rows, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	return groupItems(rows), nil
}

// ListByGrup returns the settings of one grup. An unknown or empty grup is a
// 404 — the client asked for a tab that does not exist.
func (s *PengaturanService) ListByGrup(ctx context.Context, grup string) ([]dto.SettingItem, error) {
	rows, err := s.repo.ListByGrup(ctx, grup)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("grup pengaturan %q: %w", grup, apperr.ErrNotFound)
	}
	return toItems(rows), nil
}

// Publik returns the is_publik rows as a flat kunci → cast value map, for the
// unauthenticated app shell (club name, timezone). The is_publik filter is
// applied by the query, never here.
func (s *PengaturanService) Publik(ctx context.Context) (dto.PublicSettings, error) {
	rows, err := s.repo.ListPublik(ctx)
	if err != nil {
		return nil, err
	}
	out := make(dto.PublicSettings, len(rows))
	for _, r := range rows {
		out[r.Kunci] = castOut(r.TipeNilai, r.Nilai)
	}
	return out, nil
}

// Update applies a bulk change and returns the settings as they now stand.
//
// Everything is validated before anything is written, and the write itself is
// one transaction, so a rejected key leaves the whole page untouched. Rules:
//
//   - an unknown kunci is a 422 naming that key;
//   - a value that does not fit tipe_nilai (or a non-null opsi list) is a 422
//     naming that key;
//   - a row with is_terkunci may only be changed by a super admin — checked
//     here rather than in middleware, because the route gate cannot know which
//     rows a request touches.
//
// One audit row is written per successful call, not per key: the user performed
// one action.
func (s *PengaturanService) Update(ctx context.Context, actor Actor, req dto.UpdatePengaturanReq) ([]dto.SettingGroup, error) {
	current, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	byKunci := make(map[string]domain.Pengaturan, len(current))
	for _, row := range current {
		byKunci[row.Kunci] = row
	}

	var (
		fieldErrs = map[string]string{}
		updates   = map[string]string{}
		lama      = map[string]any{}
		baru      = map[string]any{}
	)

	for _, item := range req.Items {
		row, ok := byKunci[item.Kunci]
		if !ok {
			fieldErrs[item.Kunci] = "kunci pengaturan tidak dikenal"
			continue
		}
		if row.IsTerkunci && !actor.IsSuper {
			// Reported as a field error so the client can flag the exact
			// control; the handler turns the whole call into a 403.
			fieldErrs[item.Kunci] = "hanya super admin yang boleh mengubah setelan terkunci"
			continue
		}

		var raw json.RawMessage
		if item.Nilai != nil {
			raw = *item.Nilai
		}
		nilai, err := normalizeIn(row.TipeNilai, raw)
		if err != nil {
			fieldErrs[item.Kunci] = err.Error()
			continue
		}
		if err := checkOpsi(row.Opsi, nilai); err != nil {
			fieldErrs[item.Kunci] = err.Error()
			continue
		}

		// Skip no-ops so an unchanged control does not stamp modified_by or
		// pad the audit snapshot.
		if row.Nilai != nil && *row.Nilai == nilai {
			continue
		}
		updates[item.Kunci] = nilai
		lama[item.Kunci] = castOut(row.TipeNilai, row.Nilai)
		baru[item.Kunci] = castOut(row.TipeNilai, &nilai)
	}

	if len(fieldErrs) > 0 {
		if lockedOnly(fieldErrs) {
			return nil, fmt.Errorf("setelan terkunci hanya bisa diubah super admin: %w", apperr.ErrForbidden)
		}
		return nil, apperr.Invalid(fieldErrs)
	}

	if len(updates) > 0 {
		if err := s.repo.BulkUpsert(ctx, updates, actor.UserID); err != nil {
			return nil, err
		}
		s.audit.Log(ctx, audit.Entry{
			AktorUserID: actor.UserID,
			Modul:       "pengaturan",
			Aksi:        "ubah",
			ReffType:    "mst_pengaturan",
			Ringkasan:   fmt.Sprintf("Mengubah %d setelan", len(updates)),
			NilaiLama:   lama,
			NilaiBaru:   baru,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
		})
	}

	return s.List(ctx)
}

// lockedOnly reports whether every rejection was the locked-row rule, which is
// a 403 rather than a 422 — the input was well-formed, the caller was not
// allowed to send it.
func lockedOnly(fieldErrs map[string]string) bool {
	const locked = "hanya super admin yang boleh mengubah setelan terkunci"
	for _, msg := range fieldErrs {
		if msg != locked {
			return false
		}
	}
	return len(fieldErrs) > 0
}

// toItem casts one row for the wire.
func toItem(r domain.Pengaturan) dto.SettingItem {
	return dto.SettingItem{
		ID:          r.ID,
		Kunci:       r.Kunci,
		Grup:        r.Grup,
		Label:       r.Label,
		Nilai:       castOut(r.TipeNilai, r.Nilai),
		TipeNilai:   string(r.TipeNilai),
		NilaiBawaan: castOut(r.TipeNilai, r.NilaiBawaan),
		Opsi:        r.Opsi,
		Satuan:      r.Satuan,
		Keterangan:  r.Keterangan,
		Urutan:      r.Urutan,
		IsPublik:    r.IsPublik,
		IsTerkunci:  r.IsTerkunci,
		ModifiedAt:  r.ModifiedAt,
		ModifiedBy:  r.ModifiedBy,
	}
}

func toItems(rows []domain.Pengaturan) []dto.SettingItem {
	out := make([]dto.SettingItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, toItem(r))
	}
	return out
}

// groupItems folds the ordered rows into groups, preserving the query's order
// both between groups and inside one.
func groupItems(rows []domain.Pengaturan) []dto.SettingGroup {
	groups := make([]dto.SettingGroup, 0, 6)
	index := map[string]int{}
	for _, r := range rows {
		i, ok := index[r.Grup]
		if !ok {
			groups = append(groups, dto.SettingGroup{Grup: r.Grup})
			i = len(groups) - 1
			index[r.Grup] = i
		}
		groups[i].Items = append(groups[i].Items, toItem(r))
	}
	return groups
}
