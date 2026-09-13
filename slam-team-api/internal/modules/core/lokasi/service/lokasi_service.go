// Package service contains the business rules for the lokasi module.
// Rules enforced: unique kode, coordinate/radius bounds, IANA timezone,
// foto existence, audit trail, and actor attribution.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/lokasi/domain"
	"slam-team-api/internal/modules/core/lokasi/dto"
	"slam-team-api/internal/modules/core/lokasi/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    *int64
	IPAddress string
	UserAgent string
}

// LokasiService composes repository + audit for business operations.
type LokasiService struct {
	repo    *repository.LokasiRepository
	auditor *audit.Writer
}

// NewLokasiService creates a service bound to the repository and audit writer.
func NewLokasiService(repo *repository.LokasiRepository, auditor *audit.Writer) *LokasiService {
	return &LokasiService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, inserts a row, writes an audit log,
// and returns the populated DTO.
func (s *LokasiService) Create(ctx context.Context, req dto.CreateLokasiReq, actor Actor) (*dto.LokasiResp, error) {
	// 1. Uniqueness — kode must not exist among active rows.
	exists, err := s.repo.ExistsKode(ctx, req.Kode, 0)
	if err != nil {
		return nil, fmt.Errorf("cek kode: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("kode '%s' sudah digunakan: %w", req.Kode, apperr.ErrConflict)
	}

	// 2. Timezone must be a valid IANA zone name.
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		return nil, fmt.Errorf("timezone '%s' tidak valid: %w", req.Timezone, apperr.ErrValidation)
	}

	// 3. Foto existence check.
	if req.FotoFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 4. Default is_aktif = true.
	isAktif := true
	if req.IsAktif != nil {
		isAktif = *req.IsAktif
	}

	// 5. Build domain entity.
	l := domain.Lokasi{
		Kode:        req.Kode,
		Nama:        req.Nama,
		JenisLokasi: req.JenisLokasi,
		Alamat:      req.Alamat,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		RadiusMeter: req.RadiusMeter,
		Timezone:    req.Timezone,
		FotoFileID:  req.FotoFileID,
		Keterangan:  req.Keterangan,
		IsAktif:     isAktif,
	}
	l.CreatedBy = actor.UserID

	// 6. Persist.
	if err := s.repo.Create(ctx, &l); err != nil {
		return nil, err
	}

	// 7. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "lokasi",
		Aksi:        "create",
		ReffType:    "lokasi",
		ReffID:      &l.ID,
		Ringkasan:   fmt.Sprintf("Membuat lokasi '%s'", l.Nama),
		NilaiBaru:   l,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	// 8. Fetch back with joins (foto UUID) for the response.
	created, err := s.repo.FindByID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── Update ──

// Update validates, applies changes, writes an audit log with nilai_lama/nilai_baru,
// and returns the updated DTO.
func (s *LokasiService) Update(ctx context.Context, id int64, req dto.UpdateLokasiReq, actor Actor) (*dto.LokasiResp, error) {
	// 1. Load existing row.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Uniqueness — kode must not conflict with another row.
	exists, err := s.repo.ExistsKode(ctx, req.Kode, id)
	if err != nil {
		return nil, fmt.Errorf("cek kode: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("kode '%s' sudah digunakan: %w", req.Kode, apperr.ErrConflict)
	}

	// 3. Timezone must be a valid IANA zone name.
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		return nil, fmt.Errorf("timezone '%s' tidak valid: %w", req.Timezone, apperr.ErrValidation)
	}

	// 4. Foto existence check.
	if req.FotoFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 5. Snapshot for audit.
	before := *existing

	// 6. Apply changes.
	isAktif := existing.IsAktif
	if req.IsAktif != nil {
		isAktif = *req.IsAktif
	}

	existing.Kode = req.Kode
	existing.Nama = req.Nama
	existing.JenisLokasi = req.JenisLokasi
	existing.Alamat = req.Alamat
	existing.Latitude = req.Latitude
	existing.Longitude = req.Longitude
	existing.RadiusMeter = req.RadiusMeter
	existing.Timezone = req.Timezone
	existing.FotoFileID = req.FotoFileID
	existing.Keterangan = req.Keterangan
	existing.IsAktif = isAktif
	existing.ModifiedBy = actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 7. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 8. Audit log with before/after.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "lokasi",
		Aksi:        "update",
		ReffType:    "lokasi",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Mengubah lokasi '%s'", existing.Nama),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	// 9. Fetch back with joins.
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// ── Delete ──

// SoftDelete soft-deletes a lokasi and writes an audit log.
//
// jadwal.lokasi_id references mst_lokasi with ON DELETE RESTRICT; the same
// rule is applied to the soft delete here: a lokasi still used by an active
// (non-deleted) jadwal is refused with 409 so schedules never lose their
// location.
func (s *LokasiService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	// 1. Load existing row (triggers ErrNotFound if missing).
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Usage guard.
	n, err := s.repo.CountJadwalByLokasi(ctx, id)
	if err != nil {
		return fmt.Errorf("cek pemakaian jadwal: %w", err)
	}
	if n > 0 {
		return fmt.Errorf("lokasi masih dipakai %d jadwal: %w", n, apperr.ErrConflict)
	}

	// 3. Soft-delete.
	if err := s.repo.SoftDelete(ctx, id, actor.UserID); err != nil {
		return err
	}

	// 4. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "lokasi",
		Aksi:        "delete",
		ReffType:    "lokasi",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus lokasi '%s'", existing.Nama),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return nil
}

// ── Read-only ──

// List returns a paginated list of lokasi with foto UUID.
func (s *LokasiService) List(ctx context.Context, query dto.ListLokasiQuery) ([]dto.LokasiResp, int64, error) {
	// Normalize pagination defaults.
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
		Q:       query.Q,
		IsAktif: query.IsAktif,
		Sort:    query.Sort,
		Offset:  offset,
		Limit:   perPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.LokasiResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// FindByID returns a single lokasi as DTO.
func (s *LokasiService) FindByID(ctx context.Context, id int64) (*dto.LokasiResp, error) {
	l, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(l), nil
}

// ── Mapping ──

func toResp(l *domain.Lokasi) *dto.LokasiResp {
	var modStr *string
	if l.ModifiedAt != nil {
		s := l.ModifiedAt.Format(time.RFC3339)
		modStr = &s
	}
	return &dto.LokasiResp{
		ID:          l.ID,
		Kode:        l.Kode,
		Nama:        l.Nama,
		JenisLokasi: l.JenisLokasi,
		Alamat:      l.Alamat,
		Latitude:    l.Latitude,
		Longitude:   l.Longitude,
		RadiusMeter: l.RadiusMeter,
		Timezone:    l.Timezone,
		FotoFileID:  l.FotoFileID,
		FotoUUID:    l.FotoUUID,
		Keterangan:  l.Keterangan,
		IsAktif:     l.IsAktif,
		JumlahJadwal: l.JumlahJadwal,
		CreatedAt:    l.CreatedAt.Format(time.RFC3339),
		ModifiedAt:   modStr,
	}
}
