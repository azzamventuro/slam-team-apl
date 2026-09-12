// Package service contains the business rules for the anggota module.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/anggota/domain"
	"slam-team-api/internal/modules/core/anggota/dto"
	"slam-team-api/internal/modules/core/anggota/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    *int64
	IPAddress string
	UserAgent string
}

// AnggotaService composes repository + audit for business operations.
type AnggotaService struct {
	repo    *repository.AnggotaRepository
	auditor *audit.Writer
}

// NewAnggotaService creates a service bound to the repository and audit writer.
func NewAnggotaService(repo *repository.AnggotaRepository, auditor *audit.Writer) *AnggotaService {
	return &AnggotaService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, inserts a row, writes an audit log,
// and returns the populated DTO.
func (s *AnggotaService) Create(ctx context.Context, req dto.CreateAnggotaReq, actor Actor) (*dto.AnggotaResp, error) {
	// 1. Validate instansi FK exists.
	ok, err := s.repo.InstansiExists(ctx, req.InstansiID)
	if err != nil {
		return nil, fmt.Errorf("cek instansi: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("instansi tidak ditemukan: %w", apperr.ErrValidation)
	}

	// 2. Validate wilayah FK if provided.
	if req.WilayahID != nil {
		ok, err := s.repo.WilayahExists(ctx, *req.WilayahID)
		if err != nil {
			return nil, fmt.Errorf("cek wilayah: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("wilayah tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 3. File existence checks.
	if req.FotoProfilFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoProfilFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto profil: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto profil tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.FotoFormalFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoFormalFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto formal: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto formal tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.FileIdentitasFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FileIdentitasFileID)
		if err != nil {
			return nil, fmt.Errorf("cek file identitas: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("file identitas tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 4. Parse tanggal_lahir (required).
	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		return nil, fmt.Errorf("format tanggal_lahir tidak valid: %w", apperr.ErrValidation)
	}

	// 5. Parse tanggal_bergabung (optional).
	var tglBergabung *time.Time
	if req.TanggalBergabung != nil && *req.TanggalBergabung != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalBergabung)
		if err != nil {
			return nil, fmt.Errorf("format tanggal_bergabung tidak valid: %w", apperr.ErrValidation)
		}
		tglBergabung = &t
	}

	// 6. Default status = aktif.
	statusAnggota := domain.StatusAktif
	if req.StatusAnggota != nil {
		statusAnggota = domain.StatusAnggota(*req.StatusAnggota)
	}

	// 7. Build domain entity. no_induk is ALWAYS null on create.
	a := domain.Anggota{
		InstansiID:          req.InstansiID,
		WilayahID:           req.WilayahID,
		NamaLengkap:         req.NamaLengkap,
		NamaPanggilan:       req.NamaPanggilan,
		FotoProfilFileID:    req.FotoProfilFileID,
		FotoFormalFileID:    req.FotoFormalFileID,
		JenisAnggota:        domain.JenisAnggota(req.JenisAnggota),
		JenisKelamin:        req.JenisKelamin,
		JenisIdentitas:      req.JenisIdentitas,
		NoIdentitas:         req.NoIdentitas,
		FileIdentitasFileID: req.FileIdentitasFileID,
		Pekerjaan:           req.Pekerjaan,
		Alamat:              req.Alamat,
		KodePos:             req.KodePos,
		TempatLahir:         req.TempatLahir,
		TanggalLahir:        tglLahir,
		TanggalBergabung:    tglBergabung,
		StatusAnggota:       statusAnggota,
	}
	a.CreatedBy = actor.UserID

	// 8. Persist.
	if err := s.repo.Create(ctx, &a); err != nil {
		return nil, err
	}

	// 9. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "anggota",
		Aksi:        "create",
		ReffType:    "anggota",
		ReffID:      &a.ID,
		Ringkasan:   fmt.Sprintf("Membuat anggota '%s'", a.NamaLengkap),
		NilaiBaru:   a,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	// 10. Fetch back with joins for the response.
	created, err := s.repo.FindByID(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── Update ──

// Update validates, applies changes, writes an audit log, and returns the updated DTO.
func (s *AnggotaService) Update(ctx context.Context, id int64, req dto.UpdateAnggotaReq, actor Actor) (*dto.AnggotaResp, error) {
	// 1. Load existing row.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Validate instansi FK exists.
	ok, err := s.repo.InstansiExists(ctx, req.InstansiID)
	if err != nil {
		return nil, fmt.Errorf("cek instansi: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("instansi tidak ditemukan: %w", apperr.ErrValidation)
	}

	// 3. Validate wilayah FK if provided.
	if req.WilayahID != nil {
		ok, err := s.repo.WilayahExists(ctx, *req.WilayahID)
		if err != nil {
			return nil, fmt.Errorf("cek wilayah: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("wilayah tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 4. File existence checks.
	if req.FotoProfilFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoProfilFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto profil: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto profil tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.FotoFormalFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoFormalFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto formal: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto formal tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.FileIdentitasFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FileIdentitasFileID)
		if err != nil {
			return nil, fmt.Errorf("cek file identitas: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("file identitas tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 5. Parse tanggal_lahir (required).
	tglLahir, err := time.Parse("2006-01-02", req.TanggalLahir)
	if err != nil {
		return nil, fmt.Errorf("format tanggal_lahir tidak valid: %w", apperr.ErrValidation)
	}

	// 6. Parse tanggal_bergabung (optional).
	var tglBergabung *time.Time
	if req.TanggalBergabung != nil && *req.TanggalBergabung != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalBergabung)
		if err != nil {
			return nil, fmt.Errorf("format tanggal_bergabung tidak valid: %w", apperr.ErrValidation)
		}
		tglBergabung = &t
	}

	// 7. Snapshot for audit.
	before := *existing

	// 8. Apply changes. no_induk is NEVER set from this module.
	statusAnggota := domain.StatusAnggota(existing.StatusAnggota)
	if req.StatusAnggota != nil {
		statusAnggota = domain.StatusAnggota(*req.StatusAnggota)
	}

	existing.InstansiID = req.InstansiID
	existing.WilayahID = req.WilayahID
	existing.NamaLengkap = req.NamaLengkap
	existing.NamaPanggilan = req.NamaPanggilan
	existing.FotoProfilFileID = req.FotoProfilFileID
	existing.FotoFormalFileID = req.FotoFormalFileID
	existing.JenisAnggota = domain.JenisAnggota(req.JenisAnggota)
	existing.JenisKelamin = req.JenisKelamin
	existing.JenisIdentitas = req.JenisIdentitas
	existing.NoIdentitas = req.NoIdentitas
	existing.FileIdentitasFileID = req.FileIdentitasFileID
	existing.Pekerjaan = req.Pekerjaan
	existing.Alamat = req.Alamat
	existing.KodePos = req.KodePos
	existing.TempatLahir = req.TempatLahir
	existing.TanggalLahir = tglLahir
	existing.TanggalBergabung = tglBergabung
	existing.StatusAnggota = statusAnggota
	existing.ModifiedBy = actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 9. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 10. Audit log with before/after.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "anggota",
		Aksi:        "update",
		ReffType:    "anggota",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Mengubah anggota '%s'", existing.NamaLengkap),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	// 11. Fetch back with joins.
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// ── Delete ──

// SoftDelete checks guard conditions, then soft-deletes and writes an audit log.
func (s *AnggotaService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	// 1. Load existing row (triggers ErrNotFound if missing).
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Guard: cannot delete if active KTA exists.
	hasKta, err := s.repo.HasActiveKta(ctx, id)
	if err != nil {
		return fmt.Errorf("cek kta aktif: %w", err)
	}
	if hasKta {
		return fmt.Errorf("anggota masih memiliki KTA aktif: %w", apperr.ErrConflict)
	}

	// 3. Guard: cannot delete if user account exists.
	hasUser, err := s.repo.HasUserAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("cek akun user: %w", err)
	}
	if hasUser {
		return fmt.Errorf("anggota masih memiliki akun user: %w", apperr.ErrConflict)
	}

	// 4. Soft-delete.
	if err := s.repo.SoftDelete(ctx, id, actor.UserID); err != nil {
		return err
	}

	// 5. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "anggota",
		Aksi:        "delete",
		ReffType:    "anggota",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus anggota '%s'", existing.NamaLengkap),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return nil
}

// ── Read-only ──

// List returns a paginated list of anggota with joined fields.
func (s *AnggotaService) List(ctx context.Context, query dto.ListAnggotaQuery) ([]dto.AnggotaResp, int64, error) {
	page, perPage := query.Page, query.PerPage
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	offset := (page - 1) * perPage

	q := repository.ListQuery{
		Q:          query.Q,
		InstansiID: query.InstansiID,
		Jenis:      query.Jenis,
		Status:     query.Status,
		Sort:       query.Sort,
		Offset:     offset,
		Limit:      perPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.AnggotaResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// FindByID returns a single anggota as DTO.
func (s *AnggotaService) FindByID(ctx context.Context, id int64) (*dto.AnggotaResp, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(a), nil
}

// ── Mapping ──

func toResp(a *domain.Anggota) *dto.AnggotaResp {
	var tglLahirStr string
	if !a.TanggalLahir.IsZero() {
		tglLahirStr = a.TanggalLahir.Format("2006-01-02")
	}
	var tglBergabung *string
	if a.TanggalBergabung != nil {
		s := a.TanggalBergabung.Format("2006-01-02")
		tglBergabung = &s
	}
	var modStr *string
	if a.ModifiedAt != nil {
		s := a.ModifiedAt.Format(time.RFC3339)
		modStr = &s
	}
	return &dto.AnggotaResp{
		ID:                  a.ID,
		InstansiID:          a.InstansiID,
		InstansiNama:        a.InstansiNama,
		WilayahID:           a.WilayahID,
		WilayahNama:         a.WilayahNama,
		NoInduk:             a.NoInduk,
		NamaLengkap:         a.NamaLengkap,
		NamaPanggilan:       a.NamaPanggilan,
		FotoProfilFileID:    a.FotoProfilFileID,
		FotoProfilUUID:      a.FotoProfilUUID,
		FotoFormalFileID:    a.FotoFormalFileID,
		FotoFormalUUID:      a.FotoFormalUUID,
		JenisAnggota:        string(a.JenisAnggota),
		JenisKelamin:        a.JenisKelamin,
		JenisIdentitas:      a.JenisIdentitas,
		NoIdentitas:         a.NoIdentitas,
		FileIdentitasFileID: a.FileIdentitasFileID,
		FileIdentitasUUID:   a.FileIdentitasUUID,
		Pekerjaan:           a.Pekerjaan,
		Alamat:              a.Alamat,
		KodePos:             a.KodePos,
		TempatLahir:         a.TempatLahir,
		TanggalLahir:        tglLahirStr,
		TanggalBergabung:    tglBergabung,
		StatusAnggota:       string(a.StatusAnggota),
		CreatedAt:           a.CreatedAt.Format(time.RFC3339),
		ModifiedAt:          modStr,
	}
}
