// Package service contains the business rules for the jadwal module:
// schedule validation (clock order, IANA zone, unique kode, geofence
// snapshot), the recurrence → jadwal_sesi generator, and session
// cancellation — each with an audit trail.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"slam-team-api/internal/modules/core/jadwal/domain"
	"slam-team-api/internal/modules/core/jadwal/dto"
	"slam-team-api/internal/modules/core/jadwal/repository"
	"slam-team-api/internal/shared/apperr"
	"slam-team-api/internal/shared/audit"
)

// Actor captures the authenticated user making a request.
type Actor struct {
	UserID    *int64
	IPAddress string
	UserAgent string
}

// JadwalService composes repository + audit for business operations.
type JadwalService struct {
	repo    *repository.JadwalRepository
	auditor *audit.Writer
}

// NewJadwalService creates a service bound to the repository and audit writer.
func NewJadwalService(repo *repository.JadwalRepository, auditor *audit.Writer) *JadwalService {
	return &JadwalService{repo: repo, auditor: auditor}
}

// ── Create / Update ──

// Create validates the schedule, snapshots the lokasi geofence, inserts the
// row, writes an audit log, and returns the populated DTO.
func (s *JadwalService) Create(ctx context.Context, req dto.CreateJadwalReq, actor Actor) (*dto.JadwalResp, error) {
	exists, err := s.repo.ExistsKode(ctx, req.Kode, 0)
	if err != nil {
		return nil, fmt.Errorf("cek kode: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("kode '%s' sudah digunakan: %w", req.Kode, apperr.ErrConflict)
	}

	// Column defaults per the DBML; the request overrides what it sends.
	base := domain.Jadwal{
		RadiusMeter:       100,
		Timezone:          "Asia/Jakarta",
		PolaUlang:         domain.PolaTidakBerulang,
		IntervalUlang:     1,
		ModeAbsen:         domain.ModeMasukSaja,
		WajibAbsen:        true,
		ButuhSelfie:       true,
		ButuhLokasi:       true,
		IzinkanLuarRadius: true,
		ToleransiTelatMnt: 15,
		BukaAbsenMnt:      30,
		TutupAbsenMnt:     60,
		Status:            domain.StatusDraft,
	}
	j, err := s.applyReq(ctx, base, req)
	if err != nil {
		return nil, err
	}
	j.CreatedBy = actor.UserID

	if err := s.repo.Create(ctx, j); err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "jadwal",
		Aksi:      "create",
		ReffType:  "jadwal",
		ReffID:    &j.ID,
		Ringkasan: fmt.Sprintf("Membuat jadwal '%s' (%s)", j.Nama, j.Kode),
		NilaiBaru: j,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	created, err := s.repo.FindByID(ctx, j.ID)
	if err != nil {
		return nil, err
	}
	return toResp(created), nil
}

// Update is a full replace: every column comes from the request, with
// omitted optional fields keeping their current value.
func (s *JadwalService) Update(ctx context.Context, id int64, req dto.UpdateJadwalReq, actor Actor) (*dto.JadwalResp, error) {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsKode(ctx, req.Kode, id)
	if err != nil {
		return nil, fmt.Errorf("cek kode: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("kode '%s' sudah digunakan: %w", req.Kode, apperr.ErrConflict)
	}

	before := *existing
	j, err := s.applyReq(ctx, *existing, req)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	j.ModifiedAt = &now
	j.ModifiedBy = actor.UserID

	if err := s.repo.Update(ctx, j); err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "jadwal",
		Aksi:      "update",
		ReffType:  "jadwal",
		ReffID:    &id,
		Ringkasan: fmt.Sprintf("Mengubah jadwal '%s' (%s)", j.Nama, j.Kode),
		NilaiLama: before,
		NilaiBaru: j,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	updated, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(updated), nil
}

// applyReq validates the request against the business rules and lays it over
// base (column defaults on create, the current row on update). It returns a
// new entity; base is not mutated.
func (s *JadwalService) applyReq(ctx context.Context, base domain.Jadwal, req dto.CreateJadwalReq) (*domain.Jadwal, error) {
	j := base
	fields := map[string]string{}

	// 1. Dates.
	tanggalMulai, err := parseDate(req.TanggalMulai)
	if err != nil {
		fields["tanggal_mulai"] = "format harus YYYY-MM-DD"
	}
	var tanggalSelesai, tanggalAkhir *time.Time
	if req.TanggalSelesai != nil {
		t, err := parseDate(*req.TanggalSelesai)
		switch {
		case err != nil:
			fields["tanggal_selesai"] = "format harus YYYY-MM-DD"
		case t.Before(tanggalMulai):
			fields["tanggal_selesai"] = "tidak boleh sebelum tanggal_mulai"
		default:
			tanggalSelesai = &t
		}
	}
	if req.TanggalAkhirUlang != nil {
		t, err := parseDate(*req.TanggalAkhirUlang)
		switch {
		case err != nil:
			fields["tanggal_akhir_ulang"] = "format harus YYYY-MM-DD"
		case t.Before(tanggalMulai):
			fields["tanggal_akhir_ulang"] = "tidak boleh sebelum tanggal_mulai"
		default:
			tanggalAkhir = &t
		}
	}

	// 2. Clock order — the DB CHECK is the last line, this is the 422.
	jamMulai, errM := ParseClock(req.JamMulai)
	if errM != nil {
		fields["jam_mulai"] = errM.Error()
	}
	jamSelesai, errS := ParseClock(req.JamSelesai)
	if errS != nil {
		fields["jam_selesai"] = errS.Error()
	}
	if errM == nil && errS == nil && jamSelesai.Seconds() <= jamMulai.Seconds() {
		fields["jam_selesai"] = "harus lebih besar dari jam_mulai"
	}

	// 3. Recurrence rule.
	pola := j.PolaUlang
	if req.PolaUlang != "" {
		pola = domain.PolaUlang(req.PolaUlang)
	}
	hari := normalizeHariUlang(req.HariUlang)
	switch pola {
	case domain.PolaMingguan:
		if len(hari) == 0 {
			fields["hari_ulang"] = "wajib diisi untuk pola mingguan (0 = Minggu … 6 = Sabtu)"
		}
	case domain.PolaKustom:
		// optional weekday filter
	default:
		hari = nil // tidak_berulang / harian / bulanan never read it
	}
	interval := j.IntervalUlang
	if req.IntervalUlang != nil {
		interval = *req.IntervalUlang
	}
	if interval < 1 {
		interval = 1
	}

	// 4. Lokasi + geofence snapshot. Coordinates / radius / timezone left
	// empty are copied from the referenced lokasi so the schedule keeps its own
	// stable copy even if the lokasi row changes later.
	var lokasi *repository.LokasiSnapshot
	if req.LokasiID != nil {
		l, err := s.repo.FindLokasi(ctx, *req.LokasiID)
		switch {
		case err == nil:
			lokasi = l
		case isNotFound(err):
			fields["lokasi_id"] = "lokasi tidak ditemukan"
		default:
			return nil, fmt.Errorf("cek lokasi: %w", err)
		}
	}
	lat, lng := req.Latitude, req.Longitude
	radius := j.RadiusMeter
	if req.RadiusMeter != nil {
		radius = *req.RadiusMeter
	}
	tz := j.Timezone
	if req.Timezone != "" {
		tz = req.Timezone
	}
	if lokasi != nil {
		if lat == nil || lng == nil {
			l1, l2 := lokasi.Latitude, lokasi.Longitude
			lat, lng = &l1, &l2
		}
		if req.RadiusMeter == nil {
			radius = lokasi.RadiusMeter
		}
		if req.Timezone == "" {
			tz = lokasi.Timezone
		}
	}
	if (lat == nil) != (lng == nil) {
		fields["latitude"] = "latitude dan longitude harus diisi berpasangan"
	}
	if _, err := time.LoadLocation(tz); err != nil || tz == "" {
		fields["timezone"] = fmt.Sprintf("timezone '%s' tidak valid (nama IANA, mis. Asia/Jakarta)", tz)
	}

	// 5. Optional inorga reference must exist.
	if req.InorgaID != nil {
		ok, err := s.repo.InorgaExists(ctx, *req.InorgaID)
		if err != nil {
			return nil, fmt.Errorf("cek inorga: %w", err)
		}
		if !ok {
			fields["inorga_id"] = "inorga tidak ditemukan"
		}
	}
	// TODO(22-kegiatan): validate kegiatan_id against the kegiatan table once
	// Fase 8 creates it; until then the value is stored as sent (no FK yet).

	if err := apperr.Invalid(fields); err != nil {
		return nil, err
	}

	// 6. Lay the request over the base row.
	j.Kode = req.Kode
	j.Nama = req.Nama
	j.JenisJadwal = req.JenisJadwal
	j.Deskripsi = req.Deskripsi
	j.LokasiID = req.LokasiID
	j.Latitude = lat
	j.Longitude = lng
	j.RadiusMeter = radius
	j.Timezone = tz
	j.TanggalMulai = tanggalMulai
	j.TanggalSelesai = tanggalSelesai
	j.JamMulai = jamMulai.String()
	j.JamSelesai = jamSelesai.String()
	j.PolaUlang = pola
	j.IsBerulang = pola != domain.PolaTidakBerulang
	j.HariUlang = hari
	j.IntervalUlang = interval
	j.TanggalAkhirUlang = tanggalAkhir
	if req.ModeAbsen != "" {
		j.ModeAbsen = domain.ModeAbsen(req.ModeAbsen)
	}
	if req.WajibAbsen != nil {
		j.WajibAbsen = *req.WajibAbsen
	}
	if req.ButuhSelfie != nil {
		j.ButuhSelfie = *req.ButuhSelfie
	}
	if req.ButuhLokasi != nil {
		j.ButuhLokasi = *req.ButuhLokasi
	}
	if req.IzinkanLuarRadius != nil {
		j.IzinkanLuarRadius = *req.IzinkanLuarRadius
	}
	if req.ToleransiTelatMnt != nil {
		j.ToleransiTelatMnt = *req.ToleransiTelatMnt
	}
	if req.BukaAbsenMnt != nil {
		j.BukaAbsenMnt = *req.BukaAbsenMnt
	}
	if req.TutupAbsenMnt != nil {
		j.TutupAbsenMnt = *req.TutupAbsenMnt
	}
	j.Kuota = req.Kuota
	j.KegiatanID = req.KegiatanID
	j.InorgaID = req.InorgaID
	if req.Status != "" {
		j.Status = domain.StatusJadwal(req.Status)
	}
	return &j, nil
}

// ── Delete ──

// SoftDelete soft-deletes a schedule and writes an audit log.
//
// Existing jadwal_sesi rows are left in place (FK restrict is about hard
// deletes); they become unreachable because every session read goes through
// the non-deleted parent.
//
// TODO(18-absensi): once absensi exists, refuse (409) to delete a schedule
// that already has attendance rows, so recorded presence never loses its
// parent.
func (s *JadwalService) SoftDelete(ctx context.Context, id int64, actor Actor) error {
	existing, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.SoftDelete(ctx, id, actor.UserID); err != nil {
		return err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "jadwal",
		Aksi:      "delete",
		ReffType:  "jadwal",
		ReffID:    &id,
		Ringkasan: fmt.Sprintf("Menghapus jadwal '%s' (%s)", existing.Nama, existing.Kode),
		NilaiLama: existing,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))
	return nil
}

// ── Read-only ──

// List returns one page of schedules.
func (s *JadwalService) List(ctx context.Context, query dto.ListJadwalQuery) ([]dto.JadwalResp, int64, error) {
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

	items, total, err := s.repo.List(ctx, repository.ListQuery{
		Q:             query.Q,
		Status:        query.Status,
		LokasiID:      query.LokasiID,
		PolaUlang:     query.PolaUlang,
		TanggalDari:   query.TanggalDari,
		TanggalSampai: query.TanggalSampai,
		Sort:          query.Sort,
		Offset:        (page - 1) * perPage,
		Limit:         perPage,
	})
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.JadwalResp, len(items))
	for i := range items {
		out[i] = *toResp(&items[i])
	}
	return out, total, nil
}

// FindByID returns a single schedule as DTO.
func (s *JadwalService) FindByID(ctx context.Context, id int64) (*dto.JadwalResp, error) {
	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(j), nil
}

// ListSesi returns the sessions of one schedule (404 if the schedule is gone).
func (s *JadwalService) ListSesi(ctx context.Context, jadwalID int64, query dto.ListSesiQuery) ([]dto.SesiResp, error) {
	if _, err := s.repo.FindByID(ctx, jadwalID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByJadwal(ctx, jadwalID, repository.SesiFilter{
		Status:        query.Status,
		TanggalDari:   query.TanggalDari,
		TanggalSampai: query.TanggalSampai,
	})
	if err != nil {
		return nil, err
	}
	return toSesiResps(items), nil
}

// ── Generate sesi ──

// GenerateSesi materialises the schedule's recurrence into jadwal_sesi rows
// for [dari_tanggal, sampai_tanggal]. Local wall-clock times are converted to
// UTC in the schedule's own timezone. The insert is idempotent on
// (jadwal_id, tanggal_lokal): re-running reports the existing dates as
// `dilewati` and never duplicates. A single-date schedule yields one row.
func (s *JadwalService) GenerateSesi(ctx context.Context, jadwalID int64, req dto.GenerateSesiReq, actor Actor) (*dto.GenerateSesiResp, error) {
	j, err := s.repo.FindByID(ctx, jadwalID)
	if err != nil {
		return nil, err
	}
	if j.Status == domain.StatusDibatalkan {
		return nil, fmt.Errorf("jadwal sudah dibatalkan, sesi tidak bisa dibuat: %w", apperr.ErrConflict)
	}

	loc, err := time.LoadLocation(j.Timezone)
	if err != nil {
		return nil, apperr.Validationf("timezone jadwal '%s' tidak valid", j.Timezone)
	}

	fields := map[string]string{}

	// The schedule zone is authoritative. A differing request value is the
	// browser-zone mistake the field exists to catch — reject, don't ignore.
	if req.Timezone != "" && req.Timezone != j.Timezone {
		fields["timezone"] = fmt.Sprintf("harus sama dengan timezone jadwal (%s)", j.Timezone)
	}

	// Clocks: request overrides, schedule defaults.
	mulai, err := ParseClock(firstNonEmpty(req.MulaiLokal, j.JamMulai))
	if err != nil {
		fields["mulai_lokal"] = err.Error()
	}
	selesai, err := ParseClock(firstNonEmpty(req.SelesaiLokal, j.JamSelesai))
	if err != nil {
		fields["selesai_lokal"] = err.Error()
	}
	if len(fields) == 0 && selesai.Seconds() <= mulai.Seconds() {
		fields["selesai_lokal"] = "harus lebih besar dari mulai_lokal"
	}

	// Attendance window around mulai.
	buka, tutup := j.BukaAbsenMnt, j.TutupAbsenMnt
	if req.AbsenBukaMenitSebelum != nil {
		buka = *req.AbsenBukaMenitSebelum
	}
	if req.AbsenTutupMenitSetelah != nil {
		tutup = *req.AbsenTutupMenitSetelah
	}
	if buka+tutup <= 0 {
		fields["absen_tutup_menit_setelah"] = "jendela absen kosong: buka dan tutup tidak boleh keduanya 0"
	}

	// Date window: request overrides, schedule bounds as defaults.
	dari := j.TanggalMulai
	if req.DariTanggal != nil {
		if t, err := parseDate(*req.DariTanggal); err == nil {
			dari = t
		} else {
			fields["dari_tanggal"] = "format harus YYYY-MM-DD"
		}
	}
	var sampai time.Time
	switch {
	case req.SampaiTanggal != nil:
		if t, err := parseDate(*req.SampaiTanggal); err == nil {
			sampai = t
		} else {
			fields["sampai_tanggal"] = "format harus YYYY-MM-DD"
		}
	case !j.IsBerulang:
		sampai = j.TanggalMulai
	case j.TanggalAkhirUlang != nil:
		sampai = *j.TanggalAkhirUlang
	default:
		fields["sampai_tanggal"] = "wajib diisi untuk jadwal berulang tanpa tanggal_akhir_ulang"
	}
	if len(fields) == 0 && sampai.Before(dari) {
		fields["sampai_tanggal"] = "tidak boleh sebelum dari_tanggal"
	}

	if err := apperr.Invalid(fields); err != nil {
		return nil, err
	}

	dates := Dates(Rule{
		Pola:              j.PolaUlang,
		HariUlang:         j.HariUlang,
		Interval:          j.IntervalUlang,
		TanggalMulai:      j.TanggalMulai,
		TanggalAkhirUlang: j.TanggalAkhirUlang,
	}, dari, sampai)
	if len(dates) > maxSesiPerGenerate {
		return nil, apperr.InvalidField("sampai_tanggal",
			fmt.Sprintf("rentang menghasilkan lebih dari %d sesi; persempit rentang", maxSesiPerGenerate))
	}
	if len(dates) == 0 {
		return &dto.GenerateSesiResp{Dibuat: 0, Dilewati: 0, Sesi: []dto.SesiResp{}}, nil
	}

	rows := make([]domain.Sesi, 0, len(dates))
	keys := make([]string, 0, len(dates))
	for _, date := range dates {
		mulaiUTC := mulai.At(date, loc).UTC()
		rows = append(rows, domain.Sesi{
			JadwalID:      j.ID,
			TanggalLokal:  date,
			MulaiUTC:      mulaiUTC,
			SelesaiUTC:    selesai.At(date, loc).UTC(),
			Timezone:      j.Timezone,
			AbsenBukaUTC:  mulaiUTC.Add(-time.Duration(buka) * time.Minute),
			AbsenTutupUTC: mulaiUTC.Add(time.Duration(tutup) * time.Minute),
			Status:        domain.SesiTerjadwal,
			CreatedBy:     actor.UserID,
		})
		keys = append(keys, fmtDate(date))
	}

	created, err := s.repo.CreateBatch(ctx, rows)
	if err != nil {
		return nil, err
	}
	dibuat := len(created)
	dilewati := len(dates) - dibuat

	sesi, err := s.repo.ListByJadwal(ctx, j.ID, repository.SesiFilter{Dates: keys})
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:    "jadwal",
		Aksi:     "generate_sesi",
		ReffType: "jadwal",
		ReffID:   &j.ID,
		Ringkasan: fmt.Sprintf("Generate sesi jadwal '%s': %d dibuat, %d dilewati (%s s.d. %s)",
			j.Nama, dibuat, dilewati, fmtDate(dari), fmtDate(sampai)),
		NilaiBaru: map[string]any{
			"dibuat": dibuat, "dilewati": dilewati,
			"dari_tanggal": fmtDate(dari), "sampai_tanggal": fmtDate(sampai),
			"timezone": j.Timezone, "mulai_lokal": mulai.String(), "selesai_lokal": selesai.String(),
			"absen_buka_menit_sebelum": buka, "absen_tutup_menit_setelah": tutup,
		},
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return &dto.GenerateSesiResp{Dibuat: dibuat, Dilewati: dilewati, Sesi: toSesiResps(sesi)}, nil
}

// ── Batalkan sesi ──

// BatalkanSesi cancels one session. Sessions already cancelled, already
// finished, or already in the past are refused (409).
func (s *JadwalService) BatalkanSesi(ctx context.Context, sesiID int64, req dto.BatalkanSesiReq, actor Actor) (*dto.SesiResp, error) {
	sesi, err := s.repo.FindSesiByID(ctx, sesiID)
	if err != nil {
		return nil, err
	}
	switch {
	case sesi.Status == domain.SesiDibatalkan:
		return nil, fmt.Errorf("sesi sudah dibatalkan: %w", apperr.ErrConflict)
	case sesi.Status == domain.SesiSelesai:
		return nil, fmt.Errorf("sesi sudah selesai, tidak bisa dibatalkan: %w", apperr.ErrConflict)
	case sesi.SelesaiUTC.Before(time.Now().UTC()):
		return nil, fmt.Errorf("sesi sudah lewat, tidak bisa dibatalkan: %w", apperr.ErrConflict)
	}

	before := *sesi
	if err := s.repo.CancelSesi(ctx, sesiID, req.AlasanBatal, actor.UserID); err != nil {
		return nil, err
	}
	after, err := s.repo.FindSesiByID(ctx, sesiID)
	if err != nil {
		return nil, err
	}

	s.auditor.Log(ctx, audit.Entry{
		Modul:     "jadwal",
		Aksi:      "batal_sesi",
		ReffType:  "jadwal_sesi",
		ReffID:    &sesiID,
		Ringkasan: fmt.Sprintf("Membatalkan sesi %s (jadwal #%d): %s", fmtDate(sesi.TanggalLokal), sesi.JadwalID, req.AlasanBatal),
		NilaiLama: before,
		NilaiBaru: after,
		IPAddress: actor.IPAddress,
		UserAgent: actor.UserAgent,
	}.WithRequest(nil, actor.UserID))

	return toSesiResp(after), nil
}

// ── Mapping ──

func toResp(j *domain.Jadwal) *dto.JadwalResp {
	hari := make([]int, len(j.HariUlang))
	for i, h := range j.HariUlang {
		hari[i] = int(h)
	}
	return &dto.JadwalResp{
		ID:                j.ID,
		Kode:              j.Kode,
		Nama:              j.Nama,
		JenisJadwal:       j.JenisJadwal,
		Deskripsi:         j.Deskripsi,
		LokasiID:          j.LokasiID,
		LokasiNama:        j.LokasiNama,
		Latitude:          j.Latitude,
		Longitude:         j.Longitude,
		RadiusMeter:       j.RadiusMeter,
		Timezone:          j.Timezone,
		TanggalMulai:      fmtDate(j.TanggalMulai),
		TanggalSelesai:    fmtDatePtr(j.TanggalSelesai),
		JamMulai:          j.JamMulai,
		JamSelesai:        j.JamSelesai,
		IsBerulang:        j.IsBerulang,
		PolaUlang:         string(j.PolaUlang),
		HariUlang:         hari,
		IntervalUlang:     j.IntervalUlang,
		TanggalAkhirUlang: fmtDatePtr(j.TanggalAkhirUlang),
		ModeAbsen:         string(j.ModeAbsen),
		WajibAbsen:        j.WajibAbsen,
		ButuhSelfie:       j.ButuhSelfie,
		ButuhLokasi:       j.ButuhLokasi,
		IzinkanLuarRadius: j.IzinkanLuarRadius,
		ToleransiTelatMnt: j.ToleransiTelatMnt,
		BukaAbsenMnt:      j.BukaAbsenMnt,
		TutupAbsenMnt:     j.TutupAbsenMnt,
		Kuota:             j.Kuota,
		KegiatanID:        j.KegiatanID,
		InorgaID:          j.InorgaID,
		Status:            string(j.Status),
		JumlahSesi:        j.JumlahSesi,
		CreatedAt:         j.CreatedAt.UTC().Format(time.RFC3339),
		ModifiedAt:        fmtTimePtr(j.ModifiedAt),
	}
}

func toSesiResps(items []domain.Sesi) []dto.SesiResp {
	out := make([]dto.SesiResp, len(items))
	for i := range items {
		out[i] = *toSesiResp(&items[i])
	}
	return out
}

func toSesiResp(s *domain.Sesi) *dto.SesiResp {
	// mulai_lokal / selesai_lokal are rendered in the session's own zone so a
	// calendar shows "07:00" for a WIB session regardless of the viewer's
	// browser zone. An unloadable zone (cannot happen for rows this service
	// wrote) degrades to UTC rather than failing the read.
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		loc = time.UTC
	}
	return &dto.SesiResp{
		ID:            s.ID,
		JadwalID:      s.JadwalID,
		TanggalLokal:  fmtDate(s.TanggalLokal),
		MulaiUTC:      s.MulaiUTC.UTC().Format(time.RFC3339),
		SelesaiUTC:    s.SelesaiUTC.UTC().Format(time.RFC3339),
		Timezone:      s.Timezone,
		MulaiLokal:    s.MulaiUTC.In(loc).Format("15:04"),
		SelesaiLokal:  s.SelesaiUTC.In(loc).Format("15:04"),
		AbsenBukaUTC:  s.AbsenBukaUTC.UTC().Format(time.RFC3339),
		AbsenTutupUTC: s.AbsenTutupUTC.UTC().Format(time.RFC3339),
		Status:        string(s.Status),
		AlasanBatal:   s.AlasanBatal,
		JmlDitugaskan: s.JmlDitugaskan,
		JmlHadir:      s.JmlHadir,
		Catatan:       s.Catatan,
		CreatedAt:     s.CreatedAt.UTC().Format(time.RFC3339),
		ModifiedAt:    fmtTimePtr(s.ModifiedAt),
	}
}

func fmtDatePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := fmtDate(*t)
	return &s
}

func fmtTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func isNotFound(err error) bool { return errors.Is(err, apperr.ErrNotFound) }
