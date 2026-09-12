package service

import (
	"context"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/inorga/domain"
	"slam-team-api/internal/modules/core/inorga/dto"
	"slam-team-api/internal/modules/core/inorga/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
	"slam-team-api/internal/shared/pagination"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    int64
	IPAddress string
	UserAgent string
}

// InorgaService composes repository + audit for business operations.
type InorgaService struct {
	repo    *repository.InorgaRepository
	auditor *audit.Writer
}

func NewInorgaService(repo *repository.InorgaRepository, auditor *audit.Writer) *InorgaService {
	return &InorgaService{repo: repo, auditor: auditor}
}

// ── Create ──

func (s *InorgaService) Create(ctx context.Context, req dto.CreateInorgaReq, actor Actor) (*dto.InorgaResp, error) {
	tMulai, err := time.Parse("2006-01-02", req.TanggalMulai)
	if err != nil {
		return nil, fmt.Errorf("%w: tanggal_mulai format harus YYYY-MM-DD", apperr.ErrValidation)
	}

	// Validate file refs exist.
	if req.LogoFileID != nil {
		if ok, _ := s.repo.FileExists(*req.LogoFileID); !ok {
			return nil, fmt.Errorf("%w: referensi logo", apperr.ErrNotFound)
		}
	}
	if req.BannerFileID != nil {
		if ok, _ := s.repo.FileExists(*req.BannerFileID); !ok {
			return nil, fmt.Errorf("%w: referensi banner", apperr.ErrNotFound)
		}
	}
	if req.FileSKFileID != nil {
		if ok, _ := s.repo.FileExists(*req.FileSKFileID); !ok {
			return nil, fmt.Errorf("%w: referensi file SK", apperr.ErrNotFound)
		}
	}

	// Parse optional end date.
	var tSelesai *time.Time
	if req.TanggalSelesai != nil && *req.TanggalSelesai != "" {
		t, err := time.Parse("2006-01-02", *req.TanggalSelesai)
		if err != nil {
			return nil, fmt.Errorf("%w: tanggal_selesai format harus YYYY-MM-DD", apperr.ErrValidation)
		}
		tSelesai = &t
	}

	var kode *string
	if req.Kode != "" {
		kode = &req.Kode
	}

	ent := &domain.Inorga{
		Kode:           kode,
		Nama:           req.Nama,
		TanggalMulai:   &tMulai,
		TanggalSelesai: tSelesai,
		LogoFileID:     req.LogoFileID,
		BannerFileID:   req.BannerFileID,
		FileSKFileID:   req.FileSKFileID,
		Konten:         req.Konten,
	}
	ent.CreatedBy = &actor.UserID

	if err := s.repo.Create(ent); err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul: "inorga", Aksi: "buat",
		ReffType: "mst_inorga", ReffID: &ent.ID,
		AktorUserID: &actor.UserID,
	})

	return s.toResp(ent), nil
}

// ── FindByID ──

func (s *InorgaService) FindByID(_ context.Context, id int64) (*dto.InorgaResp, error) {
	ent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	return s.toResp(ent), nil
}

// ── List ──

func (s *InorgaService) List(_ context.Context, q dto.ListInorgaQuery) (*pagination.Paginated[dto.InorgaListItem], error) {
	q.Normalize()
	items, total, err := s.repo.List(repository.ListQuery{
		Page:    q.Page,
		PerPage: q.PerPage,
		Q:       q.Q,
		Sort:    q.Sort,
		Aktif:   q.Aktif,
	})
	if err != nil {
		return nil, err
	}

	out := make([]dto.InorgaListItem, 0, len(items))
	for i := range items {
		out = append(out, s.toListItem(&items[i]))
	}

	pg := pagination.New(out, pagination.ListQuery{Page: q.Page, PerPage: q.PerPage}, total)
	return &pg, nil
}

// ── Update ──

func (s *InorgaService) Update(ctx context.Context, id int64, req dto.UpdateInorgaReq, actor Actor) (*dto.InorgaResp, error) {
	ent, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Nama != nil {
		ent.Nama = *req.Nama
	}
	if req.Kode != nil {
		ent.Kode = req.Kode
	}
	if req.TanggalMulai != nil {
		t, err := time.Parse("2006-01-02", *req.TanggalMulai)
		if err != nil {
			return nil, fmt.Errorf("%w: tanggal_mulai format harus YYYY-MM-DD", apperr.ErrValidation)
		}
		ent.TanggalMulai = &t
	}
	if req.TanggalSelesai != nil {
		if *req.TanggalSelesai == "" {
			ent.TanggalSelesai = nil
		} else {
			t, err := time.Parse("2006-01-02", *req.TanggalSelesai)
			if err != nil {
				return nil, fmt.Errorf("%w: tanggal_selesai format harus YYYY-MM-DD", apperr.ErrValidation)
			}
			ent.TanggalSelesai = &t
		}
	}
	if req.LogoFileID != nil {
		if *req.LogoFileID == 0 {
			ent.LogoFileID = nil
		} else {
			if ok, _ := s.repo.FileExists(*req.LogoFileID); !ok {
				return nil, fmt.Errorf("%w: referensi logo", apperr.ErrNotFound)
			}
			ent.LogoFileID = req.LogoFileID
		}
	}
	if req.BannerFileID != nil {
		if *req.BannerFileID == 0 {
			ent.BannerFileID = nil
		} else {
			if ok, _ := s.repo.FileExists(*req.BannerFileID); !ok {
				return nil, fmt.Errorf("%w: referensi banner", apperr.ErrNotFound)
			}
			ent.BannerFileID = req.BannerFileID
		}
	}
	if req.FileSKFileID != nil {
		if *req.FileSKFileID == 0 {
			ent.FileSKFileID = nil
		} else {
			if ok, _ := s.repo.FileExists(*req.FileSKFileID); !ok {
				return nil, fmt.Errorf("%w: referensi file SK", apperr.ErrNotFound)
			}
			ent.FileSKFileID = req.FileSKFileID
		}
	}
	if req.Konten != nil {
		ent.Konten = req.Konten
	}
	ent.ModifiedBy = &actor.UserID

	if err := s.repo.Update(ent); err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul: "inorga", Aksi: "ubah",
		ReffType: "mst_inorga", ReffID: &ent.ID,
		AktorUserID: &actor.UserID,
	})

	return s.toResp(ent), nil
}

// ── SoftDelete ──

func (s *InorgaService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	ent, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDelete(ent.ID, actor.UserID); err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul: "inorga", Aksi: "hapus",
		ReffType: "mst_inorga", ReffID: &ent.ID,
		AktorUserID: &actor.UserID,
	})
	return nil
}

// ── mappers ──

// computeAktif derives the active status from tanggal_mulai/tanggal_selesai.
func computeAktif(mulai, selesai *time.Time) bool {
	now := time.Now()
	if mulai == nil || mulai.After(now) {
		return false
	}
	if selesai != nil && selesai.Before(now) {
		return false
	}
	return true
}

func (s *InorgaService) toResp(e *domain.Inorga) *dto.InorgaResp {
	out := &dto.InorgaResp{
		ID:             e.ID,
		Kode:           e.Kode,
		Nama:           e.Nama,
		TanggalMulai:   e.TanggalMulai,
		TanggalSelesai: e.TanggalSelesai,
		Konten:         e.Konten,
		Aktif:          computeAktif(e.TanggalMulai, e.TanggalSelesai),
		CreatedAt:      e.CreatedAt,
		ModifiedAt:     e.ModifiedAt,
	}
	if e.LogoUUID != nil && e.LogoFileID != nil {
		out.Logo = &dto.FileRef{ID: *e.LogoFileID, UUID: *e.LogoUUID}
	}
	if e.BannerUUID != nil && e.BannerFileID != nil {
		out.Banner = &dto.FileRef{ID: *e.BannerFileID, UUID: *e.BannerUUID}
	}
	if e.FileSKUUID != nil && e.FileSKFileID != nil {
		out.FileSK = &dto.FileRef{ID: *e.FileSKFileID, UUID: *e.FileSKUUID}
	}
	return out
}

func (s *InorgaService) toListItem(e *domain.Inorga) dto.InorgaListItem {
	out := dto.InorgaListItem{
		ID:             e.ID,
		Kode:           e.Kode,
		Nama:           e.Nama,
		TanggalMulai:   e.TanggalMulai,
		TanggalSelesai: e.TanggalSelesai,
		Aktif:          computeAktif(e.TanggalMulai, e.TanggalSelesai),
		CreatedAt:      e.CreatedAt,
	}
	if e.LogoUUID != nil && e.LogoFileID != nil {
		out.Logo = &dto.FileRef{ID: *e.LogoFileID, UUID: *e.LogoUUID}
	}
	if e.BannerUUID != nil && e.BannerFileID != nil {
		out.Banner = &dto.FileRef{ID: *e.BannerFileID, UUID: *e.BannerUUID}
	}
	if e.FileSKUUID != nil && e.FileSKFileID != nil {
		out.FileSK = &dto.FileRef{ID: *e.FileSKFileID, UUID: *e.FileSKUUID}
	}
	return out
}
