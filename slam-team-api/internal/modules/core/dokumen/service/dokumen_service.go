// Package service contains the business rules for the dokumen module.
//
// Rules enforced: file_uuid → file_id resolution (via FileLookup interface),
// reff_type whitelist validation, auto-format from file extension, soft-delete,
// audit trail, and actor attribution.
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	filedomain "slam-team-api/internal/modules/core/file/domain"
	"slam-team-api/internal/modules/core/dokumen/domain"
	"slam-team-api/internal/modules/core/dokumen/dto"
	"slam-team-api/internal/modules/core/dokumen/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// FileLookup is the narrow interface the service needs from the file layer to
// resolve a file UUID into the full domain.File (for id + ekstensi).
// Satisfied by *filerepository.FileRepository.
type FileLookup interface {
	FindByUUID(ctx context.Context, uuid string) (filedomain.File, error)
}

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

// DokumenService composes repository + file lookup + audit for business operations.
type DokumenService struct {
	repo     *repository.DokumenRepository
	fileRepo FileLookup
	auditor  *audit.Writer
}

// NewDokumenService creates a service bound to the repository, file lookup, and audit writer.
func NewDokumenService(repo *repository.DokumenRepository, fileRepo FileLookup, auditor *audit.Writer) *DokumenService {
	return &DokumenService{repo: repo, fileRepo: fileRepo, auditor: auditor}
}

// ── Create ──

// Create validates business rules, resolves file_uuid, inserts a row, writes
// an audit log, and returns the populated DTO.
func (s *DokumenService) Create(ctx context.Context, req dto.CreateDokumenReq, actor Actor) (*dto.DokumenResp, error) {
	// 1. Validate reff_type whitelist.
	if !domain.ValidReffTypes[req.ReffType] {
		return nil, fmt.Errorf("reff_type %q tidak valid: %w", req.ReffType, apperr.ErrValidation)
	}

	// 2. Resolve file_uuid → file_id.
	f, err := s.fileRepo.FindByUUID(ctx, req.FileUUID)
	if err != nil {
		return nil, fmt.Errorf("file: %w", err)
	}
	fileID := f.ID

	// 3. Trim inputs.
	kode := strings.TrimSpace(req.Kode)
	tipe := strings.TrimSpace(req.Tipe)
	format := strings.TrimSpace(req.Format)
	keterangan := strings.TrimSpace(req.Keterangan)

	// 4. Auto-fill format from file extension if empty.
	if format == "" && f.Ekstensi != "" {
		format = f.Ekstensi
	}

	// 5. Build domain entity.
	d := domain.Dokumen{
		FileID:   &fileID,
		ReffType: req.ReffType,
		ReffID:   req.ReffID,
		Jenis:    req.Jenis,
	}
	if kode != "" {
		d.Kode = &kode
	}
	if tipe != "" {
		d.Tipe = &tipe
	}
	if format != "" {
		d.Format = &format
	}
	if keterangan != "" {
		d.Keterangan = &keterangan
	}
	d.CreatedBy = &actor.UserID

	// 6. Persist.
	if err := s.repo.Create(ctx, &d); err != nil {
		return nil, err
	}

	// 7. Audit log.
	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "dokumen",
		Aksi:        "create",
		ReffType:    "mst_dokumen",
		ReffID:      &d.ID,
		Ringkasan:   fmt.Sprintf("Membuat dokumen '%s'", safeDeref(d.Kode)),
		NilaiBaru:   d,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	// 8. Reload with File preload for the response.
	loaded, err := s.repo.FindByID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	return toResp(loaded), nil
}

// ── Update ──

// Update validates, applies changes, optionally swaps the file reference,
// writes an audit log, and returns the updated DTO.
func (s *DokumenService) Update(ctx context.Context, id int64, req dto.UpdateDokumenReq, actor Actor) (*dto.DokumenResp, error) {
	// 1. Load existing row.
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. Snapshot for audit.
	before := *existing

	// 3. Validate reff_type if provided.
	if req.ReffType != nil {
		if !domain.ValidReffTypes[*req.ReffType] {
			return nil, fmt.Errorf("reff_type %q tidak valid: %w", *req.ReffType, apperr.ErrValidation)
		}
		existing.ReffType = *req.ReffType
	}

	// 4. Apply reff_id if provided.
	if req.ReffID != nil {
		existing.ReffID = *req.ReffID
	}

	// 5. Swap file reference if file_uuid provided.
	if req.FileUUID != nil && *req.FileUUID != "" {
		f, err := s.fileRepo.FindByUUID(ctx, *req.FileUUID)
		if err != nil {
			return nil, fmt.Errorf("file: %w", err)
		}
		existing.FileID = &f.ID

		// Auto-fill format from extension if empty.
		if req.Format == nil || strings.TrimSpace(*req.Format) == "" {
			if f.Ekstensi != "" {
				existing.Format = &f.Ekstensi
			}
		}
	}

	// 6. Apply metadata fields if provided.
	if req.Kode != nil {
		kode := strings.TrimSpace(*req.Kode)
		if kode == "" {
			existing.Kode = nil
		} else {
			existing.Kode = &kode
		}
	}
	if req.Tipe != nil {
		tipe := strings.TrimSpace(*req.Tipe)
		if tipe == "" {
			existing.Tipe = nil
		} else {
			existing.Tipe = &tipe
		}
	}
	if req.Format != nil {
		format := strings.TrimSpace(*req.Format)
		if format == "" {
			// Don't clear if file was swapped (already set above).
			if req.FileUUID == nil || *req.FileUUID == "" {
				existing.Format = nil
			}
		} else {
			existing.Format = &format
		}
	}
	if req.Jenis != nil {
		existing.Jenis = req.Jenis
	}
	if req.Keterangan != nil {
		ket := strings.TrimSpace(*req.Keterangan)
		if ket == "" {
			existing.Keterangan = nil
		} else {
			existing.Keterangan = &ket
		}
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
		Modul:       "dokumen",
		Aksi:        "update",
		ReffType:    "mst_dokumen",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Mengubah dokumen '%s'", safeDeref(existing.Kode)),
		NilaiLama:   before,
		NilaiBaru:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	// 9. Reload with File preload.
	loaded, err := s.repo.FindByID(ctx, existing.ID)
	if err != nil {
		return nil, err
	}
	return toResp(loaded), nil
}

// ── Delete ──

// SoftDelete marks the document as deleted and writes an audit log.
func (s *DokumenService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ctx, id, &actor.UserID); err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		AktorUserID: &actor.UserID,
		Modul:       "dokumen",
		Aksi:        "delete",
		ReffType:    "mst_dokumen",
		ReffID:      &existing.ID,
		Ringkasan:   fmt.Sprintf("Menghapus dokumen '%s'", safeDeref(existing.Kode)),
		NilaiLama:   existing,
		IPAddress:   actor.IPAddress,
		UserAgent:   actor.UserAgent,
	})

	return nil
}

// ── Find ──

// FindByID returns one document by id as a response DTO.
func (s *DokumenService) FindByID(ctx context.Context, id int64) (*dto.DokumenResp, error) {
	d, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(d), nil
}

// ── List ──

// List returns a paginated slice of documents.
func (s *DokumenService) List(ctx context.Context, q dto.ListDokumenQuery) ([]dto.DokumenResp, int64, error) {
	q.Normalize()

	repoQ := repository.ListQuery{
		Q:        q.Q,
		ReffType: q.ReffType,
		ReffID:   q.ReffID,
		Jenis:    q.Jenis,
		Tipe:     q.Tipe,
		Format:   q.Format,
		Sort:     q.Sort,
		Offset:   (q.Page - 1) * q.PerPage,
		Limit:    q.PerPage,
	}

	items, total, err := s.repo.List(ctx, repoQ)
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.DokumenResp, len(items))
	for i := range items {
		resp[i] = *toResp(&items[i])
	}
	return resp, total, nil
}

// ── Mappers ──

func toResp(d *domain.Dokumen) *dto.DokumenResp {
	r := &dto.DokumenResp{
		ID:         d.ID,
		Kode:       d.Kode,
		Tipe:       d.Tipe,
		Format:     d.Format,
		ReffType:   d.ReffType,
		ReffID:     d.ReffID,
		Jenis:      d.Jenis,
		Keterangan: d.Keterangan,
		CreatedAt:  d.CreatedAt,
		CreatedBy:  d.CreatedBy,
		ModifiedAt: d.ModifiedAt,
	}

	// Map file domain to FileRef (never expose mst_file.id).
	if d.File != nil {
		r.File = &dto.FileRef{
			UUID:       d.File.UUID,
			NamaAsli:   d.File.NamaAsli,
			Ekstensi:   d.File.Ekstensi,
			MimeType:   d.File.MimeType,
			UkuranByte: d.File.UkuranByte,
			IsPublik:   d.File.IsPublik,
		}
	}

	return r
}

func safeDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
