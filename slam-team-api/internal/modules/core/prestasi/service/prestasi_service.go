// Package service contains the business rules for the prestasi module.
//
// Rules enforced:
//   - anggota_id must reference an active (is_deleted=false) anggota row.
//   - File references (flyer, foto_sampul) must point to existing non-deleted mst_file rows.
//   - Scope "semua" — no per-user filter (prestasi is public-readable).
//   - All changes write audit log entries.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/prestasi/domain"
	"slam-team-api/internal/modules/core/prestasi/dto"
	"slam-team-api/internal/modules/core/prestasi/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

// PrestasiService composes repository + audit for business operations.
type PrestasiService struct {
	repo    *repository.PrestasiRepository
	auditor *audit.Writer
}

// NewPrestasiService creates a service bound to the repository and audit writer.
func NewPrestasiService(repo *repository.PrestasiRepository, auditor *audit.Writer) *PrestasiService {
	return &PrestasiService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, inserts a prestasi row, writes an audit log,
// and returns the populated DTO.
func (s *PrestasiService) Create(ctx context.Context, req dto.CreatePrestasiReq, actor Actor) (*dto.PrestasiResp, error) {
	// 1. Validate anggota exists.
	ok, err := s.repo.AnggotaExists(ctx, req.AnggotaID)
	if err != nil {
		return nil, fmt.Errorf("cek anggota: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("anggota tidak ditemukan: %w", apperr.ErrValidation)
	}

	// 2. Validate file refs.
	if req.FlyerFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FlyerFileID)
		if err != nil {
			return nil, fmt.Errorf("cek flyer: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("flyer tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.FotoSampulFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoSampulFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto sampul: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto sampul tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 3. Parse tanggal_kompetisi.
	var tanggal *time.Time
	if req.TanggalKompetisi != nil && *req.TanggalKompetisi != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalKompetisi)
		if err != nil {
			return nil, fmt.Errorf("format tanggal tidak valid: %w", apperr.ErrValidation)
		}
		tanggal = &t
	}

	// 4. Build domain entity.
	p := domain.Prestasi{
		AnggotaID:        req.AnggotaID,
		Kode:             req.Kode,
		Peringkat:        req.Peringkat,
		Tingkat:          req.Tingkat,
		JudulKompetisi:   req.JudulKompetisi,
		FlyerFileID:      req.FlyerFileID,
		TanggalKompetisi: tanggal,
		AlamatKompetisi:  req.AlamatKompetisi,
		FotoSampulFileID: req.FotoSampulFileID,
		Keterangan:       req.Keterangan,
	}
	actorID := actor.UserID
	p.CreatedBy = &actorID

	// 5. Persist.
	if err := s.repo.Create(ctx, &p); err != nil {
		return nil, err
	}

	// 6. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "prestasi",
		Aksi:        "create",
		ReffType:    "prestasi",
		ReffID:      &p.ID,
		Ringkasan:   fmt.Sprintf("Membuat prestasi '%s'", p.JudulKompetisi),
		NilaiBaru:   p,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	// 7. Fetch back with joins for the response.
	created, err := s.repo.FindByID(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── FindByID ──

// FindByID returns a single prestasi by ID.
func (s *PrestasiService) FindByID(ctx context.Context, id int64) (*dto.PrestasiResp, error) {
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(p), nil
}

// ── List ──

// List returns paginated prestasi with filters.
func (s *PrestasiService) List(ctx context.Context, query dto.ListPrestasiQuery) ([]dto.PrestasiResp, int64, error) {
	q := repository.ListQuery{
		Q:            query.Q,
		AnggotaID:    query.AnggotaID,
		Tingkat:      query.Tingkat,
		Peringkat:    query.Peringkat,
		TanggalDari:  query.TanggalDari,
		TanggalSampai: query.TanggalSampai,
		Sort:         query.Sort,
		Offset:       (query.Page - 1) * query.PerPage,
		Limit:        query.PerPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.PrestasiResp, len(items))
	for i, item := range items {
		resp[i] = *toResp(&item)
	}
	return resp, total, nil
}

// ── Update ──

// Update validates, applies changes, writes an audit log, and returns the updated DTO.
func (s *PrestasiService) Update(ctx context.Context, id int64, req dto.UpdatePrestasiReq, actor Actor) (*dto.PrestasiResp, error) {
	// 1. Load existing row.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Snapshot for audit.
	before := *existing

	// 3. Validate file refs if changed.
	if req.FlyerFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FlyerFileID)
		if err != nil {
			return nil, fmt.Errorf("cek flyer: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("flyer tidak ditemukan: %w", apperr.ErrValidation)
		}
		existing.FlyerFileID = req.FlyerFileID
	}
	if req.FotoSampulFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.FotoSampulFileID)
		if err != nil {
			return nil, fmt.Errorf("cek foto sampul: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("foto sampul tidak ditemukan: %w", apperr.ErrValidation)
		}
		existing.FotoSampulFileID = req.FotoSampulFileID
	}

	// 4. Apply mutable fields.
	if req.Kode != nil {
		existing.Kode = req.Kode
	}
	if req.Peringkat != nil {
		existing.Peringkat = req.Peringkat
	}
	if req.Tingkat != nil {
		existing.Tingkat = req.Tingkat
	}
	if req.JudulKompetisi != nil {
		existing.JudulKompetisi = *req.JudulKompetisi
	}
	if req.TanggalKompetisi != nil {
		t, err := time.Parse("2006-01-02", *req.TanggalKompetisi)
		if err != nil {
			return nil, fmt.Errorf("format tanggal tidak valid: %w", apperr.ErrValidation)
		}
		existing.TanggalKompetisi = &t
	}
	if req.AlamatKompetisi != nil {
		existing.AlamatKompetisi = req.AlamatKompetisi
	}
	if req.Keterangan != nil {
		existing.Keterangan = req.Keterangan
	}

	// 5. Set modified audit.
	actorID := actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now
	existing.ModifiedBy = &actorID

	// 6. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 7. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "prestasi",
		Aksi:        "update",
		ReffType:    "prestasi",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Mengubah prestasi '%s'", existing.JudulKompetisi),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	// 8. Fetch back with joins.
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// ── SoftDelete ──

// SoftDelete marks a prestasi row as deleted and writes an audit log.
func (s *PrestasiService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	// Load for audit snapshot.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	actorID := actor.UserID
	if err := s.repo.SoftDelete(ctx, id, &actorID); err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "prestasi",
		Aksi:        "hapus",
		ReffType:    "prestasi",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Menghapus prestasi '%s'", existing.JudulKompetisi),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	return nil
}

// ── PublicList ──

// PublicList returns a public-safe projection of prestasi records.
func (s *PrestasiService) PublicList(ctx context.Context, query dto.ListPrestasiQuery) ([]dto.PublicPrestasiItem, int64, error) {
	q := repository.ListQuery{
		Q:            query.Q,
		AnggotaID:    query.AnggotaID,
		Tingkat:      query.Tingkat,
		Peringkat:    query.Peringkat,
		TanggalDari:  query.TanggalDari,
		TanggalSampai: query.TanggalSampai,
		Sort:         query.Sort,
		Offset:       (query.Page - 1) * query.PerPage,
		Limit:        query.PerPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.PublicPrestasiItem, len(items))
	for i, item := range items {
		resp[i] = *toPublicResp(&item)
	}
	return resp, total, nil
}

// ── AnggotaTersedia ──

// AnggotaTersedia returns all active anggota for dropdowns.
func (s *PrestasiService) AnggotaTersedia(ctx context.Context) ([]dto.AnggotaOpsi, error) {
	var items []dto.AnggotaOpsi
	err := s.repo.AnggotaTersedia(ctx, &items)
	return items, err
}

// ── Mappers ──

func toResp(p *domain.Prestasi) *dto.PrestasiResp {
	r := &dto.PrestasiResp{
		ID:             p.ID,
		AnggotaID:      p.AnggotaID,
		JudulKompetisi: p.JudulKompetisi,
		Kode:           p.Kode,
		Peringkat:      p.Peringkat,
		Tingkat:        p.Tingkat,
		AlamatKompetisi: p.AlamatKompetisi,
		Keterangan:     p.Keterangan,
		CreatedAt:      p.CreatedAt.Format(time.RFC3339),
	}

	if p.ModifiedAt != nil {
		s := p.ModifiedAt.Format(time.RFC3339)
		r.ModifiedAt = &s
	}

	if p.TanggalKompetisi != nil {
		s := p.TanggalKompetisi.Format("2006-01-02")
		r.TanggalKompetisi = &s
	}

	// Anggota ref.
	r.Anggota = dto.AnggotaRef{
		ID: p.AnggotaID,
	}
	if p.AnggotaNama != nil {
		r.Anggota.NamaLengkap = *p.AnggotaNama
	}
	if p.AnggotaPanggilan != nil && *p.AnggotaPanggilan != "" {
		r.Anggota.NamaPanggilan = p.AnggotaPanggilan
	}
	if p.AnggotaNoInduk != nil {
		r.Anggota.NoInduk = p.AnggotaNoInduk
	}

	// Flyer ref.
	if p.FlyerUUID != nil && p.FlyerFileIDRef != nil {
		isPublik := false
		if p.FlyerIsPublik != nil {
			isPublik = *p.FlyerIsPublik
		}
		r.Flyer = &dto.FileRef{
			FileID:   *p.FlyerFileIDRef,
			UUID:     *p.FlyerUUID,
			IsPublik: isPublik,
		}
	}

	// Foto sampul ref.
	if p.FotoSampulUUID != nil {
		isPublik := false
		if p.FotoSampulIsPublik != nil {
			isPublik = *p.FotoSampulIsPublik
		}
		// We don't have the file_id for foto_sampul from the join — use 0.
		// The frontend uses UUID for image fetch, so file_id is informational.
		r.FotoSampul = &dto.FileRef{
			FileID:   0,
			UUID:     *p.FotoSampulUUID,
			IsPublik: isPublik,
		}
	}

	return r
}

func toPublicResp(p *domain.Prestasi) *dto.PublicPrestasiItem {
	r := &dto.PublicPrestasiItem{
		JudulKompetisi: p.JudulKompetisi,
		Peringkat:      p.Peringkat,
		Tingkat:        p.Tingkat,
		AlamatKompetisi: p.AlamatKompetisi,
		FlyerUUID:      p.FlyerUUID,
		FotoSampulUUID: p.FotoSampulUUID,
	}

	if p.TanggalKompetisi != nil {
		s := p.TanggalKompetisi.Format("2006-01-02")
		r.TanggalKompetisi = &s
	}

	r.Anggota = dto.AnggotaRef{
		ID: p.AnggotaID,
	}
	if p.AnggotaNama != nil {
		r.Anggota.NamaLengkap = *p.AnggotaNama
	}
	if p.AnggotaPanggilan != nil && *p.AnggotaPanggilan != "" {
		r.Anggota.NamaPanggilan = p.AnggotaPanggilan
	}
	if p.AnggotaNoInduk != nil {
		r.Anggota.NoInduk = p.AnggotaNoInduk
	}

	return r
}
