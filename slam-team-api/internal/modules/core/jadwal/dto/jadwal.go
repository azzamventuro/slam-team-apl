// Package dto defines the request/response shapes for the jadwal module.
package dto

// CreateJadwalReq is the JSON body for POST /jadwal.
//
// Dates are "YYYY-MM-DD"; clock fields accept "HH:MM" or "HH:MM:SS" (checked in
// the service, which also enforces jam_selesai > jam_mulai). latitude /
// longitude / radius_meter / timezone left empty are snapshotted from the
// referenced lokasi. is_berulang is derived from pola_ulang, not accepted.
type CreateJadwalReq struct {
	Kode        string `json:"kode"         binding:"required,max=50"`
	Nama        string `json:"nama"         binding:"required,max=150"`
	JenisJadwal string `json:"jenis_jadwal" binding:"omitempty,max=50"`
	Deskripsi   string `json:"deskripsi"`

	LokasiID    *int64   `json:"lokasi_id"    binding:"omitempty,gt=0"`
	Latitude    *float64 `json:"latitude"     binding:"omitempty,latitude"`
	Longitude   *float64 `json:"longitude"    binding:"omitempty,longitude"`
	RadiusMeter *int     `json:"radius_meter" binding:"omitempty,min=1,max=100000"`
	Timezone    string   `json:"timezone"     binding:"omitempty,max=50"`

	TanggalMulai   string  `json:"tanggal_mulai"   binding:"required,datetime=2006-01-02"`
	TanggalSelesai *string `json:"tanggal_selesai" binding:"omitempty,datetime=2006-01-02"`
	JamMulai       string  `json:"jam_mulai"       binding:"required"`
	JamSelesai     string  `json:"jam_selesai"     binding:"required"`

	PolaUlang         string  `json:"pola_ulang"          binding:"omitempty,oneof=tidak_berulang harian mingguan bulanan kustom"`
	HariUlang         []int   `json:"hari_ulang"          binding:"omitempty,dive,min=0,max=6"`
	IntervalUlang     *int    `json:"interval_ulang"      binding:"omitempty,min=1,max=365"`
	TanggalAkhirUlang *string `json:"tanggal_akhir_ulang" binding:"omitempty,datetime=2006-01-02"`

	ModeAbsen         string `json:"mode_absen"          binding:"omitempty,oneof=masuk_saja masuk_pulang"`
	WajibAbsen        *bool  `json:"wajib_absen"         binding:"omitempty"`
	ButuhSelfie       *bool  `json:"butuh_selfie"        binding:"omitempty"`
	ButuhLokasi       *bool  `json:"butuh_lokasi"        binding:"omitempty"`
	IzinkanLuarRadius *bool  `json:"izinkan_luar_radius" binding:"omitempty"`
	ToleransiTelatMnt *int   `json:"toleransi_telat_mnt" binding:"omitempty,min=0,max=1440"`
	BukaAbsenMnt      *int   `json:"buka_absen_mnt"      binding:"omitempty,min=0,max=1440"`
	TutupAbsenMnt     *int   `json:"tutup_absen_mnt"     binding:"omitempty,min=0,max=1440"`
	Kuota             *int   `json:"kuota"               binding:"omitempty,min=0"`

	KegiatanID *int64 `json:"kegiatan_id" binding:"omitempty,gt=0"`
	InorgaID   *int64 `json:"inorga_id"   binding:"omitempty,gt=0"`
	Status     string `json:"status"      binding:"omitempty,oneof=draft terbit dibatalkan selesai"`
}

// UpdateJadwalReq is the JSON body for PUT /jadwal/:id — full replace, same shape.
type UpdateJadwalReq = CreateJadwalReq

// GenerateSesiReq is the JSON body for POST /jadwal/:id/generate-sesi
// (api-endpoints §6). Every field is optional: omitted values fall back to the
// schedule's own tanggal_mulai / tanggal_akhir_ulang / jam_* / *_absen_mnt, so
// an empty body materialises the schedule exactly as defined.
//
// timezone, when sent, must equal the schedule's timezone — conversion always
// uses jadwal.timezone, and a differing value is rejected rather than silently
// ignored, because sending the browser zone here is the exact bug the field
// exists to catch.
type GenerateSesiReq struct {
	DariTanggal            *string `json:"dari_tanggal"              binding:"omitempty,datetime=2006-01-02"`
	SampaiTanggal          *string `json:"sampai_tanggal"            binding:"omitempty,datetime=2006-01-02"`
	Timezone               string  `json:"timezone"                  binding:"omitempty,max=50"`
	MulaiLokal             string  `json:"mulai_lokal"               binding:"omitempty"`
	SelesaiLokal           string  `json:"selesai_lokal"             binding:"omitempty"`
	AbsenBukaMenitSebelum  *int    `json:"absen_buka_menit_sebelum"  binding:"omitempty,min=0,max=1440"`
	AbsenTutupMenitSetelah *int    `json:"absen_tutup_menit_setelah" binding:"omitempty,min=0,max=1440"`
}

// GenerateSesiResp is the `data` payload of generate-sesi.
type GenerateSesiResp struct {
	Dibuat   int        `json:"dibuat"`
	Dilewati int        `json:"dilewati"`
	Sesi     []SesiResp `json:"sesi"`
}

// BatalkanSesiReq is the JSON body for PATCH /sesi/:id/batalkan.
type BatalkanSesiReq struct {
	AlasanBatal string `json:"alasan_batal" binding:"required,max=1000"`
}

// JadwalResp is the standard response shape for a single jadwal.
type JadwalResp struct {
	ID          int64    `json:"id"`
	Kode        string   `json:"kode"`
	Nama        string   `json:"nama"`
	JenisJadwal string   `json:"jenis_jadwal"`
	Deskripsi   string   `json:"deskripsi"`
	LokasiID    *int64   `json:"lokasi_id"`
	LokasiNama  *string  `json:"lokasi_nama"`
	Latitude    *float64 `json:"latitude"`
	Longitude   *float64 `json:"longitude"`
	RadiusMeter int      `json:"radius_meter"`
	Timezone    string   `json:"timezone"`

	TanggalMulai   string  `json:"tanggal_mulai"`
	TanggalSelesai *string `json:"tanggal_selesai"`
	JamMulai       string  `json:"jam_mulai"`
	JamSelesai     string  `json:"jam_selesai"`

	IsBerulang        bool    `json:"is_berulang"`
	PolaUlang         string  `json:"pola_ulang"`
	HariUlang         []int   `json:"hari_ulang"`
	IntervalUlang     int     `json:"interval_ulang"`
	TanggalAkhirUlang *string `json:"tanggal_akhir_ulang"`

	ModeAbsen         string `json:"mode_absen"`
	WajibAbsen        bool   `json:"wajib_absen"`
	ButuhSelfie       bool   `json:"butuh_selfie"`
	ButuhLokasi       bool   `json:"butuh_lokasi"`
	IzinkanLuarRadius bool   `json:"izinkan_luar_radius"`
	ToleransiTelatMnt int    `json:"toleransi_telat_mnt"`
	BukaAbsenMnt      int    `json:"buka_absen_mnt"`
	TutupAbsenMnt     int    `json:"tutup_absen_mnt"`
	Kuota             *int   `json:"kuota"`

	KegiatanID *int64 `json:"kegiatan_id"`
	InorgaID   *int64 `json:"inorga_id"`
	Status     string `json:"status"`
	JumlahSesi int64  `json:"jumlah_sesi"`

	CreatedAt  string  `json:"created_at"`
	ModifiedAt *string `json:"modified_at"`
}

// SesiResp is the response shape for one jadwal_sesi row. *_utc fields are
// RFC 3339 in UTC; mulai_lokal / selesai_lokal are the same instants rendered
// in the session's own timezone ("07:00") so a calendar never has to convert.
type SesiResp struct {
	ID            int64   `json:"id"`
	JadwalID      int64   `json:"jadwal_id"`
	TanggalLokal  string  `json:"tanggal_lokal"`
	MulaiUTC      string  `json:"mulai_utc"`
	SelesaiUTC    string  `json:"selesai_utc"`
	Timezone      string  `json:"timezone"`
	MulaiLokal    string  `json:"mulai_lokal"`
	SelesaiLokal  string  `json:"selesai_lokal"`
	AbsenBukaUTC  string  `json:"absen_buka_utc"`
	AbsenTutupUTC string  `json:"absen_tutup_utc"`
	Status        string  `json:"status"`
	AlasanBatal   *string `json:"alasan_batal"`
	JmlDitugaskan int     `json:"jml_ditugaskan"`
	JmlHadir      int     `json:"jml_hadir"`
	Catatan       *string `json:"catatan"`
	CreatedAt     string  `json:"created_at"`
	ModifiedAt    *string `json:"modified_at"`
}

// ListJadwalQuery is the query-string shape for GET /jadwal.
//
// tanggal_dari / tanggal_sampai select schedules whose active span
// (tanggal_mulai … tanggal_akhir_ulang|tanggal_selesai|tanggal_mulai)
// overlaps the requested range — what a calendar month view needs.
type ListJadwalQuery struct {
	Page    int    `form:"page,default=1"      binding:"min=1"`
	PerPage int    `form:"per_page,default=20" binding:"min=1,max=100"`
	Q       string `form:"q"`
	Sort    string `form:"sort"`

	Status        string  `form:"status"         binding:"omitempty,oneof=draft terbit dibatalkan selesai"`
	LokasiID      *int64  `form:"lokasi_id"      binding:"omitempty,gt=0"`
	PolaUlang     string  `form:"pola_ulang"     binding:"omitempty,oneof=tidak_berulang harian mingguan bulanan kustom"`
	TanggalDari   *string `form:"tanggal_dari"   binding:"omitempty,datetime=2006-01-02"`
	TanggalSampai *string `form:"tanggal_sampai" binding:"omitempty,datetime=2006-01-02"`
}

// ListSesiQuery is the optional query-string filter for GET /jadwal/:id/sesi.
type ListSesiQuery struct {
	Status        string  `form:"status"         binding:"omitempty,oneof=terjadwal berlangsung selesai dibatalkan"`
	TanggalDari   *string `form:"tanggal_dari"   binding:"omitempty,datetime=2006-01-02"`
	TanggalSampai *string `form:"tanggal_sampai" binding:"omitempty,datetime=2006-01-02"`
}
