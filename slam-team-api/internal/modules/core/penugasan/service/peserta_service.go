// Package service contains the business rules for participant assignment:
// the account-aware wajib_absen rule, duplicate skipping, the notifikasi
// fan-out to account holders, the jml_ditugaskan counter, and the
// assignee-only response — each with an audit trail.
package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	notifdomain "slam-team-api/internal/modules/core/notifikasi/domain"
	notifservice "slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/internal/modules/core/penugasan/domain"
	"slam-team-api/internal/modules/core/penugasan/dto"
	"slam-team-api/internal/modules/core/penugasan/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"

	"gorm.io/gorm"
)

// Actor captures the authenticated user making a request. AnggotaID is the
// member record behind the account (JWT claim) — the key for own-row checks.
type Actor struct {
	UserID    *int64
	AnggotaID *int64
	IPAddress string
	UserAgent string
}

// PesertaService composes the repository, the notifikasi producer and the
// audit writer.
type PesertaService struct {
	repo    *repository.PesertaRepository
	notif   *notifservice.NotifikasiService
	auditor *audit.Writer
}

// NewPesertaService creates a service bound to its collaborators.
func NewPesertaService(repo *repository.PesertaRepository, notif *notifservice.NotifikasiService, auditor *audit.Writer) *PesertaService {
	return &PesertaService{repo: repo, notif: notif, auditor: auditor}
}

// ── Assign ──

// Assign adds one anggota to a schedule (or one of its sessions). It forces
// wajib_absen=false for account-less anggota, notifies the anggota's user
// when there is one, refreshes jml_ditugaskan, and audits — the row, the
// notification and the counter commit together.
func (s *PesertaService) Assign(ctx context.Context, jadwalID int64, req dto.AssignReq, actor Actor) (*domain.JadwalPeserta, error) {
	if actor.UserID == nil {
		return nil, fmt.Errorf("penugas tidak dikenal: %w", apperr.ErrUnauthorized)
	}
	jadwal, sesi, err := s.resolveTarget(ctx, jadwalID, req.SesiID)
	if err != nil {
		return nil, err
	}

	anggota, err := s.repo.FindAnggota(ctx, []int64{req.AnggotaID})
	if err != nil {
		return nil, err
	}
	if _, ok := anggota[req.AnggotaID]; !ok {
		return nil, apperr.InvalidField("anggota_id", "anggota tidak ditemukan")
	}

	var created domain.JadwalPeserta
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.LockJadwal(ctx, jadwalID); err != nil {
			return err
		}

		existing, err := repo.ExistingAnggota(ctx, jadwalID, []int64{req.AnggotaID})
		if err != nil {
			return err
		}
		if _, dup := existing[req.AnggotaID]; dup {
			return fmt.Errorf("anggota sudah ditugaskan pada jadwal ini: %w", apperr.ErrConflict)
		}

		akun, err := repo.AnggotaBerakun(ctx, []int64{req.AnggotaID})
		if err != nil {
			return err
		}

		created = newRow(jadwalID, req.SesiID, req.AnggotaID, req.PeranPeserta, req.WajibAbsen, req.Keterangan, *actor.UserID, akun)
		if err := repo.Create(ctx, &created); err != nil {
			return err
		}

		if userID, ok := akun[req.AnggotaID]; ok {
			pesan := assignPesan(jadwal, sesi, created.PeranPeserta, actor.UserID)
			if _, err := s.notif.WithTx(tx).Kirim(ctx, []int64{userID}, pesan); err != nil {
				return err
			}
		}
		return repo.SyncJmlDitugaskan(ctx, jadwalID)
	})
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "penugasan",
		Aksi:      "assign",
		ReffType:  "jadwal_peserta",
		ReffID:    &created.ID,
		Ringkasan: fmt.Sprintf("Menugaskan %s ke jadwal '%s' (%s)", anggota[req.AnggotaID].NamaLengkap, jadwal.Nama, jadwal.Kode),
		NilaiBaru: created,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return s.repo.FindByID(ctx, created.ID)
}

// BulkAssign adds many anggota at once with the same peran / wajib_absen.
// Anggota already assigned to the schedule are skipped and counted in
// dilewati; unknown anggota ids are a validation error (nothing is written).
func (s *PesertaService) BulkAssign(ctx context.Context, jadwalID int64, req dto.BulkAssignReq, actor Actor) (*dto.BulkAssignResp, error) {
	if actor.UserID == nil {
		return nil, fmt.Errorf("penugas tidak dikenal: %w", apperr.ErrUnauthorized)
	}
	jadwal, sesi, err := s.resolveTarget(ctx, jadwalID, req.SesiID)
	if err != nil {
		return nil, err
	}

	ids := uniqueIDs(req.AnggotaIDs)
	anggota, err := s.repo.FindAnggota(ctx, ids)
	if err != nil {
		return nil, err
	}
	var missing []string
	for _, id := range ids {
		if _, ok := anggota[id]; !ok {
			missing = append(missing, fmt.Sprint(id))
		}
	}
	if len(missing) > 0 {
		return nil, apperr.InvalidField("anggota_ids", "anggota tidak ditemukan: "+strings.Join(missing, ", "))
	}

	resp := &dto.BulkAssignResp{}
	var rows []domain.JadwalPeserta
	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.LockJadwal(ctx, jadwalID); err != nil {
			return err
		}

		existing, err := repo.ExistingAnggota(ctx, jadwalID, ids)
		if err != nil {
			return err
		}
		akun, err := repo.AnggotaBerakun(ctx, ids)
		if err != nil {
			return err
		}

		rows = rows[:0]
		var penerima []int64
		for _, id := range ids {
			if _, dup := existing[id]; dup {
				resp.Dilewati++
				continue
			}
			rows = append(rows, newRow(jadwalID, req.SesiID, id, req.PeranPeserta, req.WajibAbsen, nil, *actor.UserID, akun))
			if userID, ok := akun[id]; ok {
				penerima = append(penerima, userID)
			}
		}
		if len(rows) == 0 {
			return nil
		}
		if err := repo.BulkCreate(ctx, rows); err != nil {
			return err
		}
		resp.Ditugaskan = len(rows)

		pesan := assignPesan(jadwal, sesi, rows[0].PeranPeserta, actor.UserID)
		if _, err := s.notif.WithTx(tx).Kirim(ctx, penerima, pesan); err != nil {
			return err
		}
		return repo.SyncJmlDitugaskan(ctx, jadwalID)
	})
	if err != nil {
		return nil, err
	}

	if resp.Ditugaskan > 0 {
		s.auditor.Log(ctx, audit.Entry{
			Modul:     "penugasan",
			Aksi:      "assign_bulk",
			ReffType:  "jadwal",
			ReffID:    &jadwalID,
			Ringkasan: fmt.Sprintf("Menugaskan %d anggota ke jadwal '%s' (%s), %d dilewati", resp.Ditugaskan, jadwal.Nama, jadwal.Kode, resp.Dilewati),
			NilaiBaru: rows,
			IPAddress: actor.IPAddress,
			UserAgent: actor.UserAgent,
		}.WithRequest(nil, actor.UserID))
	}
	return resp, nil
}

// ── Read / Remove ──

// List returns the live assignments of a schedule.
func (s *PesertaService) List(ctx context.Context, jadwalID int64, q dto.ListPesertaQuery) ([]domain.JadwalPeserta, error) {
	if _, err := s.repo.FindJadwal(ctx, jadwalID); err != nil {
		return nil, err
	}
	items, err := s.repo.List(ctx, jadwalID, repository.ListFilter{
		SesiID:      q.SesiID,
		StatusTugas: q.StatusTugas,
		WajibAbsen:  q.WajibAbsen,
		Q:           strings.TrimSpace(q.Q),
	})
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []domain.JadwalPeserta{}
	}
	return items, nil
}

// Remove soft-deletes an anggota's assignment(s) on a schedule and refreshes
// the counter.
func (s *PesertaService) Remove(ctx context.Context, jadwalID, anggotaID int64, actor Actor) error {
	jadwal, err := s.repo.FindJadwal(ctx, jadwalID)
	if err != nil {
		return err
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.LockJadwal(ctx, jadwalID); err != nil {
			return err
		}
		n, err := repo.SoftDeleteByAnggota(ctx, jadwalID, anggotaID, actor.UserID)
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("%w: peserta", apperr.ErrNotFound)
		}
		return repo.SyncJmlDitugaskan(ctx, jadwalID)
	})
	if err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "penugasan",
		Aksi:      "remove",
		ReffType:  "jadwal",
		ReffID:    &jadwalID,
		Ringkasan: fmt.Sprintf("Menghapus anggota #%d dari jadwal '%s' (%s)", anggotaID, jadwal.Nama, jadwal.Kode),
		NilaiLama: map[string]any{"jadwal_id": jadwalID, "anggota_id": anggotaID},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))
	return nil
}

// ── Respon ──

// Respon records the assignee's accept/decline. Only the anggota behind the
// caller's account may respond to a row: anyone else — including admins —
// gets 403, because responding is a personal act, not a management one.
func (s *PesertaService) Respon(ctx context.Context, id int64, req dto.ResponReq, actor Actor) (*domain.JadwalPeserta, error) {
	row, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.AnggotaID == nil || *actor.AnggotaID != row.AnggotaID {
		return nil, fmt.Errorf("hanya peserta yang bersangkutan yang dapat merespon: %w", apperr.ErrForbidden)
	}

	jadwal, err := s.repo.FindJadwal(ctx, row.JadwalID)
	if err != nil {
		return nil, err
	}
	if jadwal.Status == "dibatalkan" || jadwal.Status == "selesai" {
		return nil, fmt.Errorf("jadwal sudah %s: %w", jadwal.Status, apperr.ErrConflict)
	}

	status := domain.StatusTugas(req.StatusTugas)
	if err := s.repo.Respon(ctx, id, status, req.Keterangan); err != nil {
		return nil, err
	}

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Tell the assigner — best effort, after the response is saved, so a
	// notification failure never blocks the response itself.
	nama := "Seorang peserta"
	if updated.AnggotaNama != nil {
		nama = *updated.AnggotaNama
	}
	verb, warna := "menerima", notifdomain.WarnaSukses
	if status == domain.TugasDitolak {
		verb, warna = "menolak", notifdomain.WarnaPeringatan
	}
	_, _ = s.notif.Kirim(ctx, []int64{updated.DitugaskanOleh}, notifservice.Pesan{
		Tipe:      "penugasan_direspon",
		Judul:     fmt.Sprintf("%s %s penugasan", nama, verb),
		Isi:       fmt.Sprintf("%s %s penugasan pada jadwal '%s' (%s).", nama, verb, jadwal.Nama, jadwal.Kode),
		Warna:     warna,
		Route:     fmt.Sprintf("/jadwal/%d", jadwal.ID),
		ReffType:  "jadwal_peserta",
		ReffID:    &updated.ID,
		CreatedBy: actor.UserID,
	})

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "penugasan",
		Aksi:      "respon",
		ReffType:  "jadwal_peserta",
		ReffID:    &id,
		Ringkasan: fmt.Sprintf("%s penugasan pada jadwal '%s' (%s)", strings.ToUpper(verb[:1])+verb[1:], jadwal.Nama, jadwal.Kode),
		NilaiLama: row,
		NilaiBaru: updated,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return updated, nil
}

// ── Helpers ──

// resolveTarget loads the schedule (404), refuses cancelled/finished ones
// (409), and — when a session is named — checks it belongs to the schedule
// and is not cancelled (422).
func (s *PesertaService) resolveTarget(ctx context.Context, jadwalID int64, sesiID *int64) (*repository.JadwalRingkas, *repository.SesiRingkas, error) {
	jadwal, err := s.repo.FindJadwal(ctx, jadwalID)
	if err != nil {
		return nil, nil, err
	}
	if jadwal.Status == "dibatalkan" || jadwal.Status == "selesai" {
		return nil, nil, fmt.Errorf("jadwal sudah %s: %w", jadwal.Status, apperr.ErrConflict)
	}
	if sesiID == nil {
		return jadwal, nil, nil
	}
	sesi, err := s.repo.FindSesi(ctx, *sesiID)
	if err != nil {
		return nil, nil, err
	}
	if sesi.JadwalID != jadwalID {
		return nil, nil, apperr.InvalidField("sesi_id", "sesi bukan milik jadwal ini")
	}
	if sesi.Status == "dibatalkan" {
		return nil, nil, apperr.InvalidField("sesi_id", "sesi sudah dibatalkan")
	}
	return jadwal, sesi, nil
}

// newRow builds one assignment. wajib_absen is the request's value (default
// true) AND the anggota has an account — an account-less anggota can never
// check in, so the flag is forced false regardless of what was asked.
func newRow(jadwalID int64, sesiID *int64, anggotaID int64, peran string, wajib *bool, keterangan *string, oleh int64, akun map[int64]int64) domain.JadwalPeserta {
	if peran == "" {
		peran = domain.PeranPeserta
	}
	wajibAbsen := wajib == nil || *wajib
	if _, berakun := akun[anggotaID]; !berakun {
		wajibAbsen = false
	}
	return domain.JadwalPeserta{
		JadwalID:       jadwalID,
		SesiID:         sesiID,
		AnggotaID:      anggotaID,
		PeranPeserta:   &peran,
		WajibAbsen:     wajibAbsen,
		StatusTugas:    domain.TugasDitugaskan,
		DitugaskanOleh: oleh,
		Keterangan:     keterangan,
	}
}

// assignPesan is the jadwal_ditugaskan notification: route to the session
// when one is named, otherwise to the schedule.
func assignPesan(jadwal *repository.JadwalRingkas, sesi *repository.SesiRingkas, peran *string, oleh *int64) notifservice.Pesan {
	route := fmt.Sprintf("/jadwal/%d", jadwal.ID)
	isi := fmt.Sprintf("Anda ditugaskan sebagai %s pada jadwal '%s' (%s).", derefOr(peran, domain.PeranPeserta), jadwal.Nama, jadwal.Kode)
	reffType, reffID := "jadwal", jadwal.ID
	if sesi != nil {
		route = fmt.Sprintf("/jadwal/%d/sesi/%d", jadwal.ID, sesi.ID)
		isi = fmt.Sprintf("Anda ditugaskan sebagai %s pada jadwal '%s' (%s), sesi %s.", derefOr(peran, domain.PeranPeserta), jadwal.Nama, jadwal.Kode, sesi.TanggalLokal.Format("2006-01-02"))
		reffType, reffID = "jadwal_sesi", sesi.ID
	}
	return notifservice.Pesan{
		Tipe:      notifdomain.TipeJadwalDitugaskan,
		Judul:     fmt.Sprintf("Penugasan: %s", jadwal.Nama),
		Isi:       isi,
		Ikon:      "calendar",
		Warna:     notifdomain.WarnaInfo,
		Route:     route,
		ReffType:  reffType,
		ReffID:    &reffID,
		Prioritas: notifdomain.PrioritasNormal,
		CreatedBy: oleh,
	}
}

func derefOr(s *string, def string) string {
	if s == nil || *s == "" {
		return def
	}
	return *s
}

// uniqueIDs drops duplicates, keeping ascending order so the inserted rows
// and the audit snapshot are deterministic.
func uniqueIDs(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
