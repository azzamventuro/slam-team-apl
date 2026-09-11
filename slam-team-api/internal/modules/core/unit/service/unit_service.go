// Package service contains the business rules for the unit module.
//
// Rules enforced:
//   - Scope resolution: "semua" (no filter) vs "milik_sendiri" (anggota_id = actor's).
//   - Create: disetujui always false; foto UUID resolved to file_id.
//   - Update: approval transition logic (false→true sets approver, true→false clears).
//   - anggota_id cannot be changed on update (unit stays with its owner).
//   - All changes write audit log entries.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/unit/domain"
	"slam-team-api/internal/modules/core/unit/dto"
	"slam-team-api/internal/modules/core/unit/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	AnggotaID int64
	IPAddress string
	UserAgent string
}

// UnitService composes repository + audit for business operations.
type UnitService struct {
	repo    *repository.UnitRepository
	auditor *audit.Writer
}

// NewUnitService creates a service bound to the repository and audit writer.
func NewUnitService(repo *repository.UnitRepository, auditor *audit.Writer) *UnitService {
	return &UnitService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, resolves foto UUID, inserts a unit row
// with disetujui=false, writes an audit log, and returns the populated DTO.
func (s *UnitService) Create(ctx context.Context, req dto.CreateUnitReq, scope string, actor Actor) (*dto.UnitResp, error) {
	// 1. Scope enforcement: milik_sendiri forces anggota_id = actor's.
	anggotaID := req.AnggotaID
	if scope == "milik_sendiri" {
		if actor.AnggotaID == 0 {
			return nil, fmt.Errorf("akun ini tidak terhubung ke data anggota: %w", apperr.ErrForbidden)
		}
		if anggotaID != actor.AnggotaID {
			return nil, fmt.Errorf("tidak dapat membuat unit untuk anggota lain: %w", apperr.ErrForbidden)
		}
		anggotaID = actor.AnggotaID
	}

	// 2. Validate anggota exists.
	ok, err := s.repo.AnggotaExists(ctx, anggotaID)
	if err != nil {
		return nil, fmt.Errorf("cek anggota: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("anggota tidak ditemukan: %w", apperr.ErrValidation)
	}

	// 3. Resolve foto_sampul_uuid → foto_sampul_file_id.
	var fotoFileID *int64
	if req.FotoSampulUUID != nil && *req.FotoSampulUUID != "" {
		id, err := s.repo.FindFileByUUID(ctx, *req.FotoSampulUUID)
		if err != nil {
			return nil, fmt.Errorf("resolve foto sampul: %w", err)
		}
		if id == 0 {
			return nil, fmt.Errorf("foto sampul tidak ditemukan: %w", apperr.ErrValidation)
		}
		fotoFileID = &id
	}

	// 4. Build domain entity — disetujui always false on create.
	u := domain.Unit{
		AnggotaID:        anggotaID,
		Kode:             req.Kode,
		Model:            req.Model,
		Panjang:          req.Panjang,
		PanjangInbar:     req.PanjangInbar,
		Lebar:            req.Lebar,
		Berat:            req.Berat,
		BeratBB:          req.BeratBB,
		FPS:              req.FPS,
		DeskripsiWarna:   req.DeskripsiWarna,
		FotoSampulFileID: fotoFileID,
		Disetujui:        false,
	}
	actorID := actor.UserID
	u.CreatedBy = &actorID

	// 5. Persist.
	if err := s.repo.Create(ctx, &u); err != nil {
		return nil, err
	}

	// 6. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actorID,
		Modul:       "unit",
		Aksi:        "create",
		ReffType:    "unit",
		ReffID:      &u.ID,
		Ringkasan:   fmt.Sprintf("Membuat unit '%s'", u.Model),
		NilaiBaru:   u,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actorID))

	// 7. Fetch back with joins for the response.
	created, err := s.repo.FindByID(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── FindByID ──

// FindByID returns a single unit by ID, applying scope filter.
func (s *UnitService) FindByID(ctx context.Context, id int64, scope string, actor Actor) (*dto.UnitResp, error) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Scope: milik_sendiri — only own units.
	if scope == "milik_sendiri" && u.AnggotaID != actor.AnggotaID {
		return nil, fmt.Errorf("unit: %w", apperr.ErrNotFound)
	}

	return toResp(u), nil
}

// ── List ──

// List returns paginated units, applying scope and filters.
func (s *UnitService) List(ctx context.Context, query dto.ListUnitQuery, scope string, actor Actor) ([]dto.UnitResp, int64, error) {
	// Scope: milik_sendiri — force anggota_id filter.
	anggotaID := query.AnggotaID
	if scope == "milik_sendiri" {
		aid := actor.AnggotaID
		anggotaID = &aid
	}

	q := repository.ListQuery{
		Q:         query.Q,
		AnggotaID: anggotaID,
		Disetujui: query.Disetujui,
		Sort:      query.Sort,
		Offset:    (query.Page - 1) * query.PerPage,
		Limit:     query.PerPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.UnitResp, len(items))
	for i, item := range items {
		resp[i] = *toResp(&item)
	}
	return resp, total, nil
}

// ── Update ──

// Update validates, applies changes, handles approval transitions,
// writes an audit log, and returns the updated DTO.
func (s *UnitService) Update(ctx context.Context, id int64, req dto.UpdateUnitReq, scope string, actor Actor) (*dto.UnitResp, error) {
	// 1. Load existing row (with joins for snapshot).
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Scope: milik_sendiri — only own units.
	if scope == "milik_sendiri" && existing.AnggotaID != actor.AnggotaID {
		return nil, fmt.Errorf("unit: %w", apperr.ErrNotFound)
	}

	// 3. Snapshot for audit.
	before := *existing

	// 4. Resolve foto UUID if provided.
	if req.FotoSampulUUID != nil && *req.FotoSampulUUID != "" {
		id, err := s.repo.FindFileByUUID(ctx, *req.FotoSampulUUID)
		if err != nil {
			return nil, fmt.Errorf("resolve foto sampul: %w", err)
		}
		if id == 0 {
			return nil, fmt.Errorf("foto sampul tidak ditemukan: %w", apperr.ErrValidation)
		}
		existing.FotoSampulFileID = &id
	}

	// 5. Apply scalar changes (skip empty strings — preserve existing).
	if req.Kode != "" {
		existing.Kode = req.Kode
	}
	if req.Model != "" {
		existing.Model = req.Model
	}
	if req.Panjang != nil {
		existing.Panjang = req.Panjang
	}
	if req.PanjangInbar != nil {
		existing.PanjangInbar = req.PanjangInbar
	}
	if req.Lebar != nil {
		existing.Lebar = req.Lebar
	}
	if req.Berat != nil {
		existing.Berat = req.Berat
	}
	if req.BeratBB != nil {
		existing.BeratBB = req.BeratBB
	}
	if req.FPS != nil {
		existing.FPS = req.FPS
	}
	if req.DeskripsiWarna != "" {
		existing.DeskripsiWarna = req.DeskripsiWarna
	}

	// 6. Approval transition (disetujui field).
	if req.Disetujui != nil && *req.Disetujui != existing.Disetujui {
		if *req.Disetujui {
			// false → true: set approver.
			existing.Disetujui = true
			existing.DisetujuiOleh = &actor.UserID
			now := time.Now().UTC()
			existing.DisetujuiPada = &now
		} else {
			// true → false: clear approver.
			existing.Disetujui = false
			existing.DisetujuiOleh = nil
			existing.DisetujuiPada = nil
		}
	}

	// 7. Set modified fields.
	existing.ModifiedBy = &actor.UserID
	now := time.Now().UTC()
	existing.ModifiedAt = &now

	// 8. Persist.
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// 9. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "unit",
		Aksi:        "update",
		ReffType:    "unit",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Mengubah unit '%s'", existing.Model),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actor.UserID))

	// 10. Fetch back with joins for the response.
	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// ── Delete ──

// SoftDelete marks a unit as deleted with audit trail.
func (s *UnitService) SoftDelete(ctx context.Context, id int64, scope string, actor Actor) error {
	// 1. Load existing row (check existence + scope).
	existing, err := s.repo.FindByIDRaw(ctx, id)
	if err != nil {
		return err
	}

	// 2. Scope: milik_sendiri — only own units.
	if scope == "milik_sendiri" && existing.AnggotaID != actor.AnggotaID {
		return fmt.Errorf("unit: %w", apperr.ErrNotFound)
	}

	// 3. Snapshot for audit.
	before := *existing

	// 4. Soft delete.
	if err := s.repo.SoftDelete(ctx, id, &actor.UserID); err != nil {
		return err
	}

	// 5. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "unit",
		Aksi:        "delete",
		ReffType:    "unit",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus unit '%s'", existing.Model),
		NilaiLama:   before,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, &actor.UserID))

	return nil
}

// ── AnggotaTersedia ──

// AnggotaTersedia returns all active anggota for the dropdown.
func (s *UnitService) AnggotaTersedia(ctx context.Context) ([]dto.AnggotaOpsi, error) {
	var items []dto.AnggotaOpsi
	err := s.repo.AnggotaTersedia(ctx, &items)
	return items, err
}

// LookupAnggotaID returns the anggota_id for a given user_id.
// Returns 0 if the user has no linked anggota.
func (s *UnitService) LookupAnggotaID(ctx context.Context, userID int64) (int64, error) {
	var anggotaID int64
	err := s.repo.DB().WithContext(ctx).
		Model(&struct{}{}).
		Table("users").
		Select("COALESCE(anggota_id, 0)").
		Where("id = ? AND is_deleted = false", userID).
		Scan(&anggotaID).Error
	return anggotaID, err
}

// ── toResp ──

// toResp converts a domain entity to the response DTO.
func toResp(u *domain.Unit) *dto.UnitResp {
	resp := &dto.UnitResp{
		ID:               u.ID,
		AnggotaID:        u.AnggotaID,
		AnggotaNama:      u.AnggotaNama,
		AnggotaNoInduk:   u.AnggotaNoInduk,
		Kode:             u.Kode,
		Model:            u.Model,
		Panjang:          u.Panjang,
		PanjangInbar:     u.PanjangInbar,
		Lebar:            u.Lebar,
		Berat:            u.Berat,
		BeratBB:          u.BeratBB,
		FPS:              u.FPS,
		DeskripsiWarna:   u.DeskripsiWarna,
		Disetujui:        u.Disetujui,
		DisetujuiOleh:    u.DisetujuiOleh,
		DisetujuiOlehNama: u.DisetujuiOlehNama,
		CreatedAt:        u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if u.ModifiedAt != nil {
		s := u.ModifiedAt.Format("2006-01-02T15:04:05Z")
		resp.ModifiedAt = &s
	}
	if u.DisetujuiPada != nil {
		s := u.DisetujuiPada.Format("2006-01-02T15:04:05Z")
		resp.DisetujuiPada = &s
	}
	if u.FotoSampulUUID != nil && *u.FotoSampulUUID != "" {
		resp.FotoSampul = &dto.FotoSampulInfo{
			UUID: *u.FotoSampulUUID,
			URL:  fmt.Sprintf("/api/v1/files/%s/medium", *u.FotoSampulUUID),
		}
	}
	return resp
}
