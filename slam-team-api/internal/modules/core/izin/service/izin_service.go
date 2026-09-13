// Package service contains the business rules for excuse requests:
//
//   - the requester is always the anggota behind the caller's account;
//   - waktu_pulang_diminta is required for pulang_cepat and forbidden for
//     every other jenis;
//   - one open (pending/approved) request per member per session;
//   - cakupan: milik_sendiri callers only ever see their own rows;
//   - approve / tolak are one-shot (409 once decided), notify the requester
//     through the notifikasi service, and are audited.
package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"slam-team-api/internal/modules/core/izin/domain"
	"slam-team-api/internal/modules/core/izin/dto"
	"slam-team-api/internal/modules/core/izin/repository"
	notifdomain "slam-team-api/internal/modules/core/notifikasi/domain"
	notifservice "slam-team-api/internal/modules/core/notifikasi/service"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"

	"gorm.io/gorm"
)

// Actor captures the authenticated user making a request. AnggotaID is the
// member record behind the account (JWT claim) — the requester's identity
// and the key for milik_sendiri checks.
type Actor struct {
	UserID    *int64
	AnggotaID *int64
	IPAddress string
	UserAgent string
}

// IzinService composes the repository, the notifikasi producer and the
// audit writer.
type IzinService struct {
	repo    *repository.IzinRepository
	notif   *notifservice.NotifikasiService
	auditor *audit.Writer
}

// NewIzinService creates a service bound to its collaborators.
func NewIzinService(repo *repository.IzinRepository, notif *notifservice.NotifikasiService, auditor *audit.Writer) *IzinService {
	return &IzinService{repo: repo, notif: notif, auditor: auditor}
}

// ── Create ──

// Create files a request for the caller's own anggota on one session.
//
// The body carries no anggota_id, so the cakupan of izin.create never widens
// who the request is for: whatever the scope, the row belongs to the caller.
// An account without an anggota has nobody to excuse and gets 403.
func (s *IzinService) Create(ctx context.Context, req dto.CreateIzinReq, scope string, actor Actor) (*dto.IzinResp, error) {
	if actor.UserID == nil {
		return nil, fmt.Errorf("pengaju tidak dikenal: %w", apperr.ErrUnauthorized)
	}
	if actor.AnggotaID == nil {
		return nil, fmt.Errorf("akun ini tidak terhubung ke data anggota: %w", apperr.ErrForbidden)
	}
	anggotaID := *actor.AnggotaID

	// 1. Cross-field rules the binding tags cannot express.
	jenis := domain.JenisIzin(req.Jenis)
	fields := map[string]string{}
	var waktu *string
	if jenis == domain.JenisPulangCepat {
		if req.WaktuPulangDiminta == nil || strings.TrimSpace(*req.WaktuPulangDiminta) == "" {
			fields["waktu_pulang_diminta"] = "wajib diisi untuk jenis pulang_cepat"
		} else if w, err := parseWaktu(*req.WaktuPulangDiminta); err != nil {
			fields["waktu_pulang_diminta"] = err.Error()
		} else {
			waktu = &w
		}
	} else if req.WaktuPulangDiminta != nil && strings.TrimSpace(*req.WaktuPulangDiminta) != "" {
		fields["waktu_pulang_diminta"] = "hanya untuk jenis pulang_cepat"
	}
	if strings.TrimSpace(req.Alasan) == "" {
		fields["alasan"] = "wajib diisi"
	}
	if err := apperr.Invalid(fields); err != nil {
		return nil, err
	}

	// 2. The session must exist and still be something one can be excused
	// from. A cancelled session needs no excuse.
	sesi, err := s.repo.FindSesi(ctx, req.SesiID)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			return nil, apperr.InvalidField("sesi_id", "sesi tidak ditemukan")
		}
		return nil, err
	}
	if sesi.Status == "dibatalkan" {
		return nil, apperr.InvalidField("sesi_id", "sesi sudah dibatalkan")
	}

	// 3. Attachment, when named, must be a live mst_file row.
	if req.LampiranFileID != nil {
		ok, err := s.repo.FileExists(ctx, *req.LampiranFileID)
		if err != nil {
			return nil, fmt.Errorf("cek lampiran: %w", err)
		}
		if !ok {
			return nil, apperr.InvalidField("lampiran_file_id", "berkas lampiran tidak ditemukan")
		}
	}

	// 4. One open request per member per session.
	open, err := s.repo.HasOpenRequest(ctx, sesi.ID, anggotaID)
	if err != nil {
		return nil, err
	}
	if open {
		return nil, fmt.Errorf("sudah ada pengajuan izin yang menunggu/disetujui untuk sesi ini: %w", apperr.ErrConflict)
	}

	// 5. Persist.
	jadwalID := sesi.JadwalID
	z := domain.Izin{
		SesiID:             sesi.ID,
		JadwalID:           &jadwalID,
		AnggotaID:          anggotaID,
		Jenis:              jenis,
		Alasan:             strings.TrimSpace(req.Alasan),
		WaktuPulangDiminta: waktu,
		LampiranFileID:     req.LampiranFileID,
		Status:             domain.StatusMenunggu,
	}
	z.CreatedBy = actor.UserID
	if err := s.repo.Create(ctx, &z); err != nil {
		return nil, err
	}

	created, err := s.repo.FindByID(ctx, z.ID)
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "izin",
		Aksi:      "create",
		ReffType:  "absensi_izin",
		ReffID:    &created.ID,
		Ringkasan: fmt.Sprintf("Mengajukan %s untuk sesi %s jadwal '%s' (%s)", jenis, sesi.TanggalLokal.Format("2006-01-02"), sesi.JadwalNama, sesi.JadwalKode),
		NilaiBaru: created,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return toResp(created), nil
}

// ── List ──

// List returns one page of requests. The cakupan resolved by the PermGuard
// for izin.read decides the owner filter: milik_sendiri narrows to the
// caller's anggota (an account without one sees nothing); semua — and the
// super admin, for whom the guard sets no cakupan at all — sees all.
func (s *IzinService) List(ctx context.Context, q dto.ListIzinQuery, scope string, actor Actor) ([]dto.IzinResp, int64, error) {
	q.Normalize()

	var owner *int64
	if scope == "milik_sendiri" {
		if actor.AnggotaID == nil {
			return []dto.IzinResp{}, 0, nil
		}
		owner = actor.AnggotaID
	}

	sort, desc, _ := q.SortColumn("created_at", "status", "jenis", "diproses_pada")
	items, total, err := s.repo.List(ctx, repository.ListFilter{
		AnggotaID: owner,
		Status:    q.Status,
		SesiID:    q.SesiID,
		JadwalID:  q.JadwalID,
		Jenis:     q.Jenis,
		Q:         strings.TrimSpace(q.Q),
		Sort:      sort,
		Desc:      desc,
		Offset:    q.Offset(),
		Limit:     q.Limit(),
	})
	if err != nil {
		return nil, 0, err
	}

	resp := make([]dto.IzinResp, 0, len(items))
	for i := range items {
		resp = append(resp, *toResp(&items[i]))
	}
	return resp, total, nil
}

// ── Approve / Tolak ──

// Approve grants a pending request: status → disetujui, reviewer + time
// stamped, the member's roster rows flagged izin, and the requester notified
// (izin_disetujui) — all in one transaction. A request already decided is a
// 409, never silently re-approved.
func (s *IzinService) Approve(ctx context.Context, id int64, actor Actor) (*dto.IzinResp, error) {
	if actor.UserID == nil {
		return nil, fmt.Errorf("peninjau tidak dikenal: %w", apperr.ErrUnauthorized)
	}
	before, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if before.Status != domain.StatusMenunggu {
		return nil, fmt.Errorf("izin sudah %s: %w", before.Status, apperr.ErrConflict)
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		if err := repo.Approve(ctx, id, *actor.UserID); err != nil {
			return err
		}

		// The roster: an excused member is no longer "ditugaskan" on that
		// session. Only for jenis that replace attendance — a pulang_cepat
		// member still shows up — and only the session-specific assignment
		// row: a schedule-wide row (sesi_id NULL) covers every session, so
		// flipping it would excuse the member from all of them.
		if before.Jenis.MenjadiKehadiran() {
			if _, err := repo.MarkPesertaIzin(ctx, before.SesiID, before.AnggotaID); err != nil {
				return err
			}
		}

		// TODO(18-absensi): write/update the member's absensi row for this
		// session with izin_id = before.ID and status_kehadiran = jenis
		// (izin / sakit / dinas) so the session-closing job never marks them
		// alfa. The absensi table is created by the 18-absensi migration;
		// until then an approved izin is only visible through this table
		// (and the jadwal_peserta.status_tugas = 'izin' flag above).

		_, err := s.notif.WithTx(tx).KirimKeAnggota(ctx, []int64{before.AnggotaID}, keputusanPesan(before, domain.StatusDisetujui, "", actor.UserID))
		return err
	})
	if err != nil {
		return nil, err
	}

	after, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "izin",
		Aksi:      "approve",
		ReffType:  "absensi_izin",
		ReffID:    &id,
		Ringkasan: fmt.Sprintf("Menyetujui %s %s untuk sesi %s", before.Jenis, namaAnggota(before), tanggalSesi(before)),
		NilaiLama: before,
		NilaiBaru: after,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return toResp(after), nil
}

// Tolak refuses a pending request with the reviewer's note and notifies the
// requester (izin_ditolak). Same one-shot rule as Approve.
func (s *IzinService) Tolak(ctx context.Context, id int64, req dto.TolakReq, actor Actor) (*dto.IzinResp, error) {
	if actor.UserID == nil {
		return nil, fmt.Errorf("peninjau tidak dikenal: %w", apperr.ErrUnauthorized)
	}
	catatan := strings.TrimSpace(req.CatatanPeninjau)
	if catatan == "" {
		return nil, apperr.InvalidField("catatan_peninjau", "wajib diisi")
	}
	before, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if before.Status != domain.StatusMenunggu {
		return nil, fmt.Errorf("izin sudah %s: %w", before.Status, apperr.ErrConflict)
	}

	err = s.repo.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).Tolak(ctx, id, *actor.UserID, catatan); err != nil {
			return err
		}
		_, err := s.notif.WithTx(tx).KirimKeAnggota(ctx, []int64{before.AnggotaID}, keputusanPesan(before, domain.StatusDitolak, catatan, actor.UserID))
		return err
	})
	if err != nil {
		return nil, err
	}

	after, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "izin",
		Aksi:      "tolak",
		ReffType:  "absensi_izin",
		ReffID:    &id,
		Ringkasan: fmt.Sprintf("Menolak %s %s untuk sesi %s", before.Jenis, namaAnggota(before), tanggalSesi(before)),
		NilaiLama: before,
		NilaiBaru: after,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return toResp(after), nil
}

// ── Helpers ──

// keputusanPesan is the izin_disetujui / izin_ditolak notification to the
// requester, routed to their own izin list.
func keputusanPesan(z *domain.Izin, status domain.StatusIzin, catatan string, oleh *int64) notifservice.Pesan {
	sesi := tanggalSesi(z)
	jadwal := ""
	if z.JadwalNama != nil {
		jadwal = fmt.Sprintf(" jadwal '%s'", *z.JadwalNama)
	}
	p := notifservice.Pesan{
		Ikon:      "envelope-paper",
		Route:     "/izin",
		ReffType:  "absensi_izin",
		ReffID:    &z.ID,
		Prioritas: notifdomain.PrioritasNormal,
		CreatedBy: oleh,
	}
	if status == domain.StatusDisetujui {
		p.Tipe = domain.NotifTipeDisetujui
		p.Judul = fmt.Sprintf("Izin disetujui: %s", z.Jenis)
		p.Isi = fmt.Sprintf("Pengajuan %s Anda untuk sesi %s%s telah disetujui.", z.Jenis, sesi, jadwal)
		p.Warna = notifdomain.WarnaSukses
		return p
	}
	p.Tipe = domain.NotifTipeDitolak
	p.Judul = fmt.Sprintf("Izin ditolak: %s", z.Jenis)
	p.Isi = fmt.Sprintf("Pengajuan %s Anda untuk sesi %s%s ditolak. Catatan: %s", z.Jenis, sesi, jadwal, catatan)
	p.Warna = notifdomain.WarnaPeringatan
	return p
}

func namaAnggota(z *domain.Izin) string {
	if z.AnggotaNama != nil && *z.AnggotaNama != "" {
		return *z.AnggotaNama
	}
	return fmt.Sprintf("anggota #%d", z.AnggotaID)
}

func tanggalSesi(z *domain.Izin) string {
	if z.SesiTanggal != nil {
		return z.SesiTanggal.Format("2006-01-02")
	}
	return fmt.Sprintf("#%d", z.SesiID)
}

// parseWaktu accepts "HH:MM" or "HH:MM:SS" and returns the PG `time` text
// form "HH:MM:SS".
func parseWaktu(s string) (string, error) {
	parts := strings.Split(strings.TrimSpace(s), ":")
	if len(parts) != 2 && len(parts) != 3 {
		return "", fmt.Errorf("format jam harus HH:MM atau HH:MM:SS")
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || len(p) != 2 || n < 0 {
			return "", fmt.Errorf("format jam harus HH:MM atau HH:MM:SS")
		}
		nums[i] = n
	}
	if nums[0] > 23 || nums[1] > 59 || nums[2] > 59 {
		return "", fmt.Errorf("jam di luar rentang 00:00:00–23:59:59")
	}
	return fmt.Sprintf("%02d:%02d:%02d", nums[0], nums[1], nums[2]), nil
}

const tsLayout = time.RFC3339

func toResp(z *domain.Izin) *dto.IzinResp {
	r := &dto.IzinResp{
		ID:                 z.ID,
		SesiID:             z.SesiID,
		SesiStatus:         z.SesiStatus,
		JadwalID:           z.JadwalID,
		JadwalNama:         z.JadwalNama,
		JadwalKode:         z.JadwalKode,
		AnggotaID:          z.AnggotaID,
		AnggotaNama:        z.AnggotaNama,
		NoInduk:            z.NoInduk,
		Jenis:              string(z.Jenis),
		Alasan:             z.Alasan,
		WaktuPulangDiminta: z.WaktuPulangDiminta,
		Status:             string(z.Status),
		DiprosesOleh:       z.DiprosesOleh,
		DiprosesOlehNama:   z.DiprosesOlehNama,
		CatatanPeninjau:    z.CatatanPeninjau,
		CreatedAt:          z.CreatedAt.UTC().Format(tsLayout),
	}
	if z.SesiTanggal != nil {
		d := z.SesiTanggal.Format("2006-01-02")
		r.SesiTanggal = &d
	}
	if z.DiprosesPada != nil {
		t := z.DiprosesPada.UTC().Format(tsLayout)
		r.DiprosesPada = &t
	}
	if z.ModifiedAt != nil {
		t := z.ModifiedAt.UTC().Format(tsLayout)
		r.ModifiedAt = &t
	}
	if z.LampiranFileID != nil && z.LampiranUUID != nil && *z.LampiranUUID != "" {
		r.Lampiran = &dto.LampiranInfo{
			FileID: *z.LampiranFileID,
			UUID:   *z.LampiranUUID,
			URL:    fmt.Sprintf("/api/v1/files/%s/original", *z.LampiranUUID),
		}
	}
	return r
}
