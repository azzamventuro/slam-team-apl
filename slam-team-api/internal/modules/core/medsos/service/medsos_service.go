// Package service contains the business rules for the medsos module.
// Rules enforced: FK validation (anggota exists), cakupan scoping (milik_sendiri),
// soft-delete, audit trail, and actor attribution.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/medsos/domain"
	"slam-team-api/internal/modules/core/medsos/dto"
	"slam-team-api/internal/modules/core/medsos/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

// MedsosService composes repository + audit for business operations.
type MedsosService struct {
	repo    *repository.MedsosRepository
	auditor *audit.Writer
}

// NewMedsosService creates a service bound to the repository and audit writer.
func NewMedsosService(repo *repository.MedsosRepository, auditor *audit.Writer) *MedsosService {
	return &MedsosService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, inserts a row, writes an audit log,
// and returns the populated DTO.
func (s *MedsosService) Create(ctx context.Context, req dto.CreateMedsosReq, actor Actor, cakupan string) (*dto.MedsosResp, error) {
	// 1. Resolve cakupan — milik_sendiri forces anggota_id to the actor's own.
	effectiveAnggotaID := req.AnggotaID
	if cakupan == "milik_sendiri" {
		userAnggotaID, err := s.repo.GetAnggotaIDByUserID(ctx, actor.UserID)
		if err != nil || userAnggotaID == nil {
			return nil, fmt.Errorf("data anggota pengguna tidak ditemukan: %w", apperr.ErrForbidden)
		}
		if req.AnggotaID != *userAnggotaID {
			return nil, fmt.Errorf("tidak diizinkan membuat medsos anggota lain: %w", apperr.ErrForbidden)
		}
		effectiveAnggotaID = *userAnggotaID
	}

	// 2. Validate anggota exists.
	exists, err := s.repo.AnggotaExists(ctx, effectiveAnggotaID)
	if err != nil {
		return nil, fmt.Errorf("cek anggota: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("anggota tidak ditemukan: %w", apperr.ErrNotFound)
	}

	// 3. Trim inputs.
	jenisMedsos := strings.TrimSpace(req.JenisMedsos)
	kontenMedsos := strings.TrimSpace(req.KontenMedsos)
	icon := strings.TrimSpace(req.Icon)
	kode := strings.TrimSpace(req.Kode)

	// 4. Default icon from jenis_medsos if empty.
	if icon == "" {
		icon = jenisMedsos
	}

	// 5. Build domain entity.
	m := domain.Medsos{
		AnggotaID:    effectiveAnggotaID,
		JenisMedsos:  jenisMedsos,
		KontenMedsos: kontenMedsos,
		Tipe:         req.Tipe,
	}
	if icon != "" {
		m.Icon = &icon
	}
	if kode != "" {
		m.Kode = &kode
	}
	m.CreatedBy = &actor.UserID

	// 6. Persist.
	if err := s.repo.Create(ctx, &m); err != nil {
		return nil, err
	}

	// 7. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "medsos",
		Aksi:        "create",
		ReffType:    "mst_medsos",
		ReffID:      &m.ID,
		Ringkasan:   fmt.Sprintf("Membuat medsos '%s'", jenisMedsos),
		NilaiBaru:   m,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	return toResp(&m), nil
}

// ── Update ──

// Update validates, applies changes, writes an audit log, and returns the updated DTO.
func (s *MedsosService) Update(ctx context.Context, id int64, req dto.UpdateMedsosReq, actor Actor, cakupan string) (*dto.MedsosResp, error) {
	// 1. Resolve scope for find.
	var anggotaScope *int64
	if cakupan == "milik_sendiri" {
		userAnggotaID, err := s.repo.GetAnggotaIDByUserID(ctx, actor.UserID)
		if err != nil || userAnggotaID == nil {
			return nil, fmt.Errorf("data anggota pengguna tidak ditemukan: %w", apperr.ErrForbidden)
		}
		anggotaScope = userAnggotaID
	}

	// 2. Load existing row (scoped).
	existing, err := s.repo.FindByIDScoped(ctx, id, anggotaScope)
	if err != nil {
		return nil, err
	}

	// 3. Snapshot for audit.
	before := *existing

	// 4. Trim inputs.
	jenisMedsos := strings.TrimSpace(req.JenisMedsos)
	kontenMedsos := strings.TrimSpace(req.KontenMedsos)
	icon := strings.TrimSpace(req.Icon)
	kode := strings.TrimSpace(req.Kode)

	// 5. Default icon from jenis_medsos if empty.
	if icon == "" {
		icon = jenisMedsos
	}

	// 6. Apply changes.
	existing.JenisMedsos = jenisMedsos
	existing.KontenMedsos = kontenMedsos
	existing.Tipe = req.Tipe
	if icon != "" {
		existing.Icon = &icon
	} else {
		existing.Icon = nil
	}
	if kode != "" {
		existing.Kode = &kode
	} else {
		existing.Kode = nil
	}
	existing.ModifiedBy = &actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 7. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 8. Audit log with before/after.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "medsos",
		Aksi:        "update",
		ReffType:    "mst_medsos",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Mengubah medsos '%s'", jenisMedsos),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	return toResp(existing), nil
}

// ── Delete ──

// SoftDelete marks a row as deleted with cakupan scoping.
func (s *MedsosService) SoftDelete(ctx context.Context, id int64, actor Actor, cakupan string) error {
	// 1. Resolve scope for find.
	var anggotaScope *int64
	if cakupan == "milik_sendiri" {
		userAnggotaID, err := s.repo.GetAnggotaIDByUserID(ctx, actor.UserID)
		if err != nil || userAnggotaID == nil {
			return fmt.Errorf("data anggota pengguna tidak ditemukan: %w", apperr.ErrForbidden)
		}
		anggotaScope = userAnggotaID
	}

	// 2. Load existing row (scoped).
	existing, err := s.repo.FindByIDScoped(ctx, id, anggotaScope)
	if err != nil {
		return nil
	}

	// 3. Soft-delete.
	if err := s.repo.SoftDelete(ctx, id, &actor.UserID); err != nil {
		return err
	}

	// 4. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "medsos",
		Aksi:        "delete",
		ReffType:    "mst_medsos",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus medsos '%s'", existing.JenisMedsos),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	return nil
}

// ── Read-only ──

// FindByID returns a single medsos by ID (with cakupan scoping).
func (s *MedsosService) FindByID(ctx context.Context, id int64, actor Actor, cakupan string) (*dto.MedsosResp, error) {
	var anggotaScope *int64
	if cakupan == "milik_sendiri" {
		userAnggotaID, err := s.repo.GetAnggotaIDByUserID(ctx, actor.UserID)
		if err != nil || userAnggotaID == nil {
			return nil, fmt.Errorf("data anggota pengguna tidak ditemukan: %w", apperr.ErrForbidden)
		}
		anggotaScope = userAnggotaID
	}

	m, err := s.repo.FindByIDScoped(ctx, id, anggotaScope)
	if err != nil {
		return nil, err
	}
	return toResp(m), nil
}

// List returns a paginated list of medsos with cakupan scoping.
func (s *MedsosService) List(ctx context.Context, query dto.ListMedsosQuery, actor Actor, cakupan string) ([]dto.MedsosResp, int64, error) {
	// Normalize pagination.
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

	// Resolve cakupan scope.
	var anggotaScope *int64
	if cakupan == "milik_sendiri" {
		userAnggotaID, err := s.repo.GetAnggotaIDByUserID(ctx, actor.UserID)
		if err != nil || userAnggotaID == nil {
			return nil, 0, fmt.Errorf("data anggota pengguna tidak ditemukan: %w", apperr.ErrForbidden)
		}
		anggotaScope = userAnggotaID
	}

	q := repository.ListQuery{
		Q:           query.Q,
		AnggotaID:   query.AnggotaID,
		JenisMedsos: query.JenisMedsos,
		Sort:        query.Sort,
		Offset:      offset,
		Limit:       perPage,
	}

	items, total, err := s.repo.List(ctx, q, anggotaScope)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.MedsosResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// ── Mapping ──

func toResp(m *domain.Medsos) *dto.MedsosResp {
	var modStr *string
	if m.ModifiedAt != nil {
		s := m.ModifiedAt.Format(time.RFC3339)
		modStr = &s
	}
	return &dto.MedsosResp{
		ID:           m.ID,
		AnggotaID:    m.AnggotaID,
		Kode:         m.Kode,
		Tipe:         m.Tipe,
		Icon:         m.Icon,
		JenisMedsos:  m.JenisMedsos,
		KontenMedsos: m.KontenMedsos,
		CreatedAt:    m.CreatedAt.Format(time.RFC3339),
		CreatedBy:    m.CreatedBy,
		ModifiedAt:   modStr,
		ModifiedBy:   m.ModifiedBy,
	}
}

// ── AnggotaTersedia ──

// AnggotaTersedia returns all active anggota for dropdowns.
func (s *MedsosService) AnggotaTersedia(ctx context.Context) ([]dto.AnggotaOpsi, error) {
	var items []dto.AnggotaOpsi
	err := s.repo.AnggotaTersedia(ctx, &items)
	return items, err
}
