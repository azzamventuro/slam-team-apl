// Package service contains the business rules for the instansi module.
// Rules enforced: unique kode, file existence, soft-delete guard (anggota count),
// default status, audit trail, and actor attribution.
package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/instansi/domain"
	"slam-team-api/internal/modules/core/instansi/dto"
	"slam-team-api/internal/modules/core/instansi/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    *int64
	IPAddress string
	UserAgent string
}

// InstansiService composes repository + audit for business operations.
type InstansiService struct {
	repo    *repository.InstansiRepository
	auditor *audit.Writer
}

// NewInstansiService creates a service bound to the repository and audit writer.
func NewInstansiService(repo *repository.InstansiRepository, auditor *audit.Writer) *InstansiService {
	return &InstansiService{repo: repo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, inserts a row, writes an audit log,
// and returns the populated DTO.
func (s *InstansiService) Create(ctx context.Context, req dto.CreateInstansiReq, actor Actor) (*dto.InstansiResp, error) {
	// 1. Uniqueness — kode must not exist among active rows.
	exists, err := s.repo.ExistsKode(ctx, req.Kode, 0)
	if err != nil {
		return nil, fmt.Errorf("cek kode: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("kode '%s' sudah digunakan: %w", req.Kode, apperr.ErrConflict)
	}

	// 2. File existence checks.
	if req.LogoUtamaFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.LogoUtamaFileID)
		if err != nil {
			return nil, fmt.Errorf("cek logo utama: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("logo utama tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.LogoTambahanFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.LogoTambahanFileID)
		if err != nil {
			return nil, fmt.Errorf("cek logo tambahan: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("logo tambahan tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 3. Parse tanggal_bergabung (optional, format YYYY-MM-DD).
	var tglBergabung *time.Time
	if req.TanggalBergabung != nil && *req.TanggalBergabung != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalBergabung)
		if err != nil {
			return nil, fmt.Errorf("format tanggal_bergabung tidak valid: %w", apperr.ErrValidation)
		}
		tglBergabung = &t
	}

	// 4. Default status = 1 (aktif).
	status := 1
	if req.Status != nil {
		status = *req.Status
	}

	// 5. Build domain entity.
	inst := domain.Instansi{
		Kode:                req.Kode,
		Nama:                req.Nama,
		NamaClub:            req.NamaClub,
		Alamat:              req.Alamat,
		NoTelepon:           req.NoTelepon,
		LogoUtamaFileID:     req.LogoUtamaFileID,
		LogoTambahanFileID:  req.LogoTambahanFileID,
		TanggalBergabung:    tglBergabung,
		Status:              status,
	}
	inst.CreatedBy = actor.UserID

	// 6. Persist.
	if err := s.repo.Create(ctx, &inst); err != nil {
		return nil, err
	}

	// 7. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "instansi",
		Aksi:        "create",
		ReffType:    "instansi",
		ReffID:      &inst.ID,
		Ringkasan:   fmt.Sprintf("Membuat instansi '%s'", inst.Nama),
		NilaiBaru:   inst,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	// 8. Fetch back with joins (logo UUIDs, jumlah_anggota) for the response.
	created, err := s.repo.FindByID(ctx, inst.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// ── Update ──

// Update validates, applies changes, writes an audit log with nilai_lama/nilai_baru,
// and returns the updated DTO.
func (s *InstansiService) Update(ctx context.Context, id int64, req dto.UpdateInstansiReq, actor Actor) (*dto.InstansiResp, error) {
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

	// 3. File existence checks.
	if req.LogoUtamaFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.LogoUtamaFileID)
		if err != nil {
			return nil, fmt.Errorf("cek logo utama: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("logo utama tidak ditemukan: %w", apperr.ErrValidation)
		}
	}
	if req.LogoTambahanFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.LogoTambahanFileID)
		if err != nil {
			return nil, fmt.Errorf("cek logo tambahan: %w", err)
		}
		if !ok {
			return nil, fmt.Errorf("logo tambahan tidak ditemukan: %w", apperr.ErrValidation)
		}
	}

	// 4. Parse tanggal_bergabung.
	var tglBergabung *time.Time
	if req.TanggalBergabung != nil && *req.TanggalBergabung != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalBergabung)
		if err != nil {
			return nil, fmt.Errorf("format tanggal_bergabung tidak valid: %w", apperr.ErrValidation)
		}
		tglBergabung = &t
	}

	// 5. Snapshot for audit.
	before := *existing

	// 6. Apply changes.
	status := existing.Status
	if req.Status != nil {
		status = *req.Status
	}

	existing.Kode = req.Kode
	existing.Nama = req.Nama
	existing.NamaClub = req.NamaClub
	existing.Alamat = req.Alamat
	existing.NoTelepon = req.NoTelepon
	existing.LogoUtamaFileID = req.LogoUtamaFileID
	existing.LogoTambahanFileID = req.LogoTambahanFileID
	existing.TanggalBergabung = tglBergabung
	existing.Status = status
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
		Modul:       "instansi",
		Aksi:        "update",
		ReffType:    "instansi",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Mengubah instansi '%s'", existing.Nama),
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

// SoftDelete checks that no active anggota reference this instansi,
// then soft-deletes and writes an audit log.
func (s *InstansiService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	// 1. Load existing row (triggers ErrNotFound if missing).
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// 2. Guard: cannot delete while anggota still reference this instansi.
	count, err := s.repo.CountAnggotaAktif(ctx, id)
	if err != nil {
		return fmt.Errorf("cek jumlah anggota: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("instansi masih dipakai oleh %d anggota: %w", count, apperr.ErrConflict)
	}

	// 3. Soft-delete.
	if err := s.repo.SoftDelete(ctx, id, actor.UserID); err != nil {
		return err
	}

	// 4. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: actor.UserID,
		Modul:       "instansi",
		Aksi:        "delete",
		ReffType:    "instansi",
		ReffID:      &id,
		Ringkasan:   fmt.Sprintf("Menghapus instansi '%s'", existing.Nama),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return nil
}

// ── Read-only ──

// List returns a paginated list of instansi with logo UUIDs and jumlah_anggota.
func (s *InstansiService) List(ctx context.Context, query dto.ListInstansiQuery) ([]dto.InstansiResp, int64, error) {
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
		Q:      query.Q,
		Status: query.Status,
		Sort:   query.Sort,
		Offset: offset,
		Limit:  perPage,
	}

	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.InstansiResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// FindByID returns a single instansi as DTO.
func (s *InstansiService) FindByID(ctx context.Context, id int64) (*dto.InstansiResp, error) {
	inst, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(inst), nil
}

// ── Mapping ──

func toResp(inst *domain.Instansi) *dto.InstansiResp {
	var tglStr *string
	if inst.TanggalBergabung != nil {
		s := inst.TanggalBergabung.Format("2006-01-02")
		tglStr = &s
	}
	var modStr *string
	if inst.ModifiedAt != nil {
		s := inst.ModifiedAt.Format(time.RFC3339)
		modStr = &s
	}
	return &dto.InstansiResp{
		ID:                  inst.ID,
		Kode:                inst.Kode,
		Nama:                inst.Nama,
		NamaClub:            inst.NamaClub,
		Alamat:              inst.Alamat,
		NoTelepon:           inst.NoTelepon,
		LogoUtamaFileID:     inst.LogoUtamaFileID,
		LogoUtamaUUID:       inst.LogoUtamaUUID,
		LogoTambahanFileID:  inst.LogoTambahanFileID,
		LogoTambahanUUID:    inst.LogoTambahanUUID,
		TanggalBergabung:    tglStr,
		Status:              inst.Status,
		JumlahAnggota:       inst.JumlahAnggota,
		CreatedAt:           inst.CreatedAt.Format(time.RFC3339),
		ModifiedAt:          modStr,
	}
}
