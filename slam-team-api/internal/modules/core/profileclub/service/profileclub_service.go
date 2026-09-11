// Package service contains the business rules for the profile_club module.
//
// Rules enforced: singleton guard (at most one active row), file existence
// validation, public profile aggregation (prestasi), audit trail, and actor
// attribution.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/profileclub/domain"
	"slam-team-api/internal/modules/core/profileclub/dto"
	"slam-team-api/internal/modules/core/profileclub/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// FileExistChecker is the narrow interface to validate file existence.
type FileExistChecker interface {
	FileExists(ctx context.Context, id int64) (bool, error)
}

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

// ProfileClubService composes repository + file checker + audit for business operations.
type ProfileClubService struct {
	repo      *repository.ProfileClubRepository
	fileCheck FileExistChecker
	auditor   *audit.Writer
}

// NewProfileClubService creates a service bound to dependencies.
func NewProfileClubService(repo *repository.ProfileClubRepository, fileCheck FileExistChecker, auditor *audit.Writer) *ProfileClubService {
	return &ProfileClubService{repo: repo, fileCheck: fileCheck, auditor: auditor}
}

// ── Create ──

// Create inserts the club profile. Returns 409 if an active row already exists.
func (s *ProfileClubService) Create(ctx context.Context, req dto.UpsertProfileClubReq, actor Actor) (*dto.ProfileClubResp, error) {
	// 1. Singleton guard.
	count, err := s.repo.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("profil klub sudah ada, gunakan ubah: %w", apperr.ErrConflict)
	}

	// 2. Validate file IDs.
	if err := s.validateFileIDs(ctx, &req); err != nil {
		return nil, err
	}

	// 3. Build domain entity.
	d := domain.ProfileClub{
		Nama:              req.Nama,
		Singkatan:         req.Singkatan,
		BannerFileID:      req.BannerFileID,
		LogoSimpleFileID:  req.LogoSimpleFileID,
		LogoBesarFileID:   req.LogoBesarFileID,
		Alamat:            req.Alamat,
		Keterangan:        req.Keterangan,
	}
	d.CreatedBy = &actor.UserID

	// 4. Persist.
	if err := s.repo.Create(ctx, &d); err != nil {
		return nil, err
	}

	// 5. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "profile_club",
		Aksi:        "create",
		ReffType:    "profile_club",
		ReffID:      &d.ID,
		Ringkasan:   fmt.Sprintf("Membuat profil klub '%s'", d.Nama),
		NilaiBaru:   d,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	// 6. Reload with joins for UUID resolution.
	loaded, err := s.repo.FindByID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	return toResp(loaded), nil
}

// ── Update ──

// Update applies changes to the existing club profile.
func (s *ProfileClubService) Update(ctx context.Context, id int64, req dto.UpsertProfileClubReq, actor Actor) (*dto.ProfileClubResp, error) {
	// 1. Load existing.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Snapshot for audit.
	before := *existing

	// 3. Validate file IDs.
	if err := s.validateFileIDs(ctx, &req); err != nil {
		return nil, err
	}

	// 4. Apply changes.
	existing.Nama = req.Nama
	existing.Singkatan = req.Singkatan
	existing.BannerFileID = req.BannerFileID
	existing.LogoSimpleFileID = req.LogoSimpleFileID
	existing.LogoBesarFileID = req.LogoBesarFileID
	existing.Alamat = req.Alamat
	existing.Keterangan = req.Keterangan
	existing.ModifiedBy = &actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 5. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 6. Audit log with before/after.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "profile_club",
		Aksi:        "update",
		ReffType:    "profile_club",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Mengubah profil klub '%s'", existing.Nama),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	// 7. Reload with joins.
	loaded, err := s.repo.FindByID(ctx, existing.ID)
	if err != nil {
		return nil, err
	}
	return toResp(loaded), nil
}

// ── Read ──

// FindByID loads a profile_club by id.
func (s *ProfileClubService) FindByID(ctx context.Context, id int64) (*dto.ProfileClubResp, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(d), nil
}

// List returns the (practically 0/1) profile_club rows.
func (s *ProfileClubService) List(ctx context.Context, q repository.ListQuery) ([]dto.ProfileClubResp, int64, error) {
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}
	resp := make([]dto.ProfileClubResp, len(items))
	for i := range items {
		resp[i] = *toResp(&items[i])
	}
	return resp, total, nil
}

// ── Delete ──

// SoftDelete marks the profile as deleted.
func (s *ProfileClubService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, id, &actor.UserID); err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "profile_club",
		Aksi:        "delete",
		ReffType:    "profile_club",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Menghapus profil klub '%s'", existing.Nama),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	return nil
}

// ── Public ──

// GetPublicProfil returns the public-safe profile + prestasi aggregation.
func (s *ProfileClubService) GetPublicProfil(ctx context.Context) (*dto.PublicProfilResp, error) {
	resp := &dto.PublicProfilResp{Prestasi: []dto.PrestasiPublik{}}

	d, err := s.repo.GetActive(ctx)
	if err != nil {
		// If no profile exists, return empty response (landing still renders).
		if isNotFound(err) {
			return resp, nil
		}
		return nil, err
	}

	resp.Nama = d.Nama
	resp.Singkatan = d.Singkatan
	resp.Alamat = d.Alamat
	resp.Keterangan = d.Keterangan

	// Build public URLs from resolved UUIDs.
	if d.BannerUUID != nil {
		resp.BannerURL = "/api/v1/files/" + *d.BannerUUID + "/medium"
	}
	if d.LogoSimpleUUID != nil {
		resp.LogoSimpleURL = "/api/v1/files/" + *d.LogoSimpleUUID + "/low"
	}
	if d.LogoBesarUUID != nil {
		resp.LogoBesarURL = "/api/v1/files/" + *d.LogoBesarUUID + "/medium"
	}

	// Aggregated prestasi (Guest-readable).
	prestasi, err := s.repo.GetPublicPrestasi(ctx)
	if err == nil && len(prestasi) > 0 {
		resp.Prestasi = make([]dto.PrestasiPublik, len(prestasi))
		for i, p := range prestasi {
			resp.Prestasi[i] = dto.PrestasiPublik{
				JudulKompetisi: p.JudulKompetisi,
				Tingkat:        safeDeref(p.Tingkat),
				Peringkat:      safeDeref(p.Peringkat),
			}
			if p.TanggalKompetisi != nil {
				resp.Prestasi[i].TanggalKompetisi = p.TanggalKompetisi
			}
		}
	}

	return resp, nil
}

// ── Helpers ──

// validateFileIDs checks that each non-nil file ID exists in mst_file.
func (s *ProfileClubService) validateFileIDs(ctx context.Context, req *dto.UpsertProfileClubReq) error {
	type fieldRef struct {
		name string
		id   *int64
	}
	for _, f := range []fieldRef{
		{"banner", req.BannerFileID},
		{"logo simple", req.LogoSimpleFileID},
		{"logo besar", req.LogoBesarFileID},
	} {
		if f.id != nil {
			ok, err := s.fileCheck.FileExists(ctx, *f.id)
			if err != nil {
				return fmt.Errorf("cek %s: %w", f.name, err)
			}
			if !ok {
				return fmt.Errorf("%s tidak ditemukan: %w", f.name, apperr.ErrValidation)
			}
		}
	}
	return nil
}

func toResp(d *domain.ProfileClub) *dto.ProfileClubResp {
	return &dto.ProfileClubResp{
		ID:               d.ID,
		Nama:             d.Nama,
		Singkatan:        d.Singkatan,
		BannerFileID:     d.BannerFileID,
		LogoSimpleFileID: d.LogoSimpleFileID,
		LogoBesarFileID:  d.LogoBesarFileID,
		BannerUUID:       d.BannerUUID,
		LogoSimpleUUID:   d.LogoSimpleUUID,
		LogoBesarUUID:    d.LogoBesarUUID,
		Alamat:           d.Alamat,
		Keterangan:       d.Keterangan,
		CreatedAt:        d.CreatedAt,
		CreatedBy:        d.CreatedBy,
		ModifiedAt:       d.ModifiedAt,
	}
}

func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ListQueryFromDTO converts a DTO ListQuery to a repository ListQuery.
func ListQueryFromDTO(q dto.ListProfileClubQuery) repository.ListQuery {
	return repository.ListQuery{
		Q:      q.Q,
		Sort:   q.Sort,
		Offset: (q.Page - 1) * q.PerPage,
		Limit:  q.PerPage,
	}
}

func isNotFound(err error) bool {
	return errors.Is(err, apperr.ErrNotFound)
}
