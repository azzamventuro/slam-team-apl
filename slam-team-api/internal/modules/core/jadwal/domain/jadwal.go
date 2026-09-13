// Package domain holds the GORM entities for the jadwal module: the schedule
// definition (jadwal) and the concrete dated sessions materialised from its
// recurrence rule (jadwal_sesi).
package domain

import (
	"time"

	"slam-team-api/internal/shared/model"
)

// PolaUlang mirrors the pola_ulang PG enum.
type PolaUlang string

const (
	PolaTidakBerulang PolaUlang = "tidak_berulang"
	PolaHarian        PolaUlang = "harian"
	PolaMingguan      PolaUlang = "mingguan"
	PolaBulanan       PolaUlang = "bulanan"
	PolaKustom        PolaUlang = "kustom"
)

// ModeAbsen mirrors the mode_absen PG enum.
type ModeAbsen string

const (
	ModeMasukSaja   ModeAbsen = "masuk_saja"
	ModeMasukPulang ModeAbsen = "masuk_pulang"
)

// StatusJadwal is the varchar(20) status column — deliberately not a PG enum
// per the DBML (draft / terbit / dibatalkan / selesai).
type StatusJadwal string

const (
	StatusDraft      StatusJadwal = "draft"
	StatusTerbit     StatusJadwal = "terbit"
	StatusDibatalkan StatusJadwal = "dibatalkan"
	StatusSelesai    StatusJadwal = "selesai"
)

// StatusSesi mirrors the status_sesi PG enum.
type StatusSesi string

const (
	SesiTerjadwal   StatusSesi = "terjadwal"
	SesiBerlangsung StatusSesi = "berlangsung"
	SesiSelesai     StatusSesi = "selesai"
	SesiDibatalkan  StatusSesi = "dibatalkan"
)

// Jadwal maps to the jadwal table — a schedule definition with a recurrence
// rule and the attendance rules every session inherits. latitude / longitude /
// radius_meter are a snapshot of the referenced mst_lokasi taken when the
// schedule is created, so old attendance rules stay stable if the location
// row changes later.
//
// JamMulai / JamSelesai are the PG `time` columns kept as "HH:MM:SS" strings:
// there is no date to attach them to until generate-sesi combines them with a
// tanggal_lokal in the schedule's own timezone.
type Jadwal struct {
	ID                int64        `gorm:"column:id;primaryKey"          json:"id"`
	Kode              string       `gorm:"column:kode"                   json:"kode"`
	Nama              string       `gorm:"column:nama"                   json:"nama"`
	JenisJadwal       string       `gorm:"column:jenis_jadwal"           json:"jenis_jadwal"`
	Deskripsi         string       `gorm:"column:deskripsi"              json:"deskripsi"`
	LokasiID          *int64       `gorm:"column:lokasi_id"              json:"lokasi_id,omitempty"`
	Latitude          *float64     `gorm:"column:latitude"               json:"latitude,omitempty"`
	Longitude         *float64     `gorm:"column:longitude"              json:"longitude,omitempty"`
	RadiusMeter       int          `gorm:"column:radius_meter"           json:"radius_meter"`
	Timezone          string       `gorm:"column:timezone"               json:"timezone"`
	TanggalMulai      time.Time    `gorm:"column:tanggal_mulai;type:date" json:"tanggal_mulai"`
	TanggalSelesai    *time.Time   `gorm:"column:tanggal_selesai;type:date" json:"tanggal_selesai,omitempty"`
	JamMulai          string       `gorm:"column:jam_mulai;type:time"    json:"jam_mulai"`
	JamSelesai        string       `gorm:"column:jam_selesai;type:time"  json:"jam_selesai"`
	IsBerulang        bool         `gorm:"column:is_berulang"            json:"is_berulang"`
	PolaUlang         PolaUlang    `gorm:"column:pola_ulang"             json:"pola_ulang"`
	HariUlang         IntArray     `gorm:"column:hari_ulang;type:int[]"  json:"hari_ulang"`
	IntervalUlang     int          `gorm:"column:interval_ulang"         json:"interval_ulang"`
	TanggalAkhirUlang *time.Time   `gorm:"column:tanggal_akhir_ulang;type:date" json:"tanggal_akhir_ulang,omitempty"`
	ModeAbsen         ModeAbsen    `gorm:"column:mode_absen"             json:"mode_absen"`
	WajibAbsen        bool         `gorm:"column:wajib_absen"            json:"wajib_absen"`
	ButuhSelfie       bool         `gorm:"column:butuh_selfie"           json:"butuh_selfie"`
	ButuhLokasi       bool         `gorm:"column:butuh_lokasi"           json:"butuh_lokasi"`
	IzinkanLuarRadius bool         `gorm:"column:izinkan_luar_radius"    json:"izinkan_luar_radius"`
	ToleransiTelatMnt int          `gorm:"column:toleransi_telat_mnt"    json:"toleransi_telat_mnt"`
	BukaAbsenMnt      int          `gorm:"column:buka_absen_mnt"         json:"buka_absen_mnt"`
	TutupAbsenMnt     int          `gorm:"column:tutup_absen_mnt"        json:"tutup_absen_mnt"`
	Kuota             *int         `gorm:"column:kuota"                  json:"kuota,omitempty"`
	KegiatanID        *int64       `gorm:"column:kegiatan_id"            json:"kegiatan_id,omitempty"`
	InorgaID          *int64       `gorm:"column:inorga_id"              json:"inorga_id,omitempty"`
	Status            StatusJadwal `gorm:"column:status"                 json:"status"`

	// Read-only projections resolved by the repository's list/detail query,
	// never written back (GORM "->" permission).
	LokasiNama *string `gorm:"column:lokasi_nama;->" json:"lokasi_nama,omitempty"`
	JumlahSesi int64   `gorm:"column:jumlah_sesi;->" json:"jumlah_sesi"`

	model.Audit
}

// TableName pins the table name for GORM.
func (Jadwal) TableName() string { return "jadwal" }

// Sesi maps to jadwal_sesi — one concrete date of a schedule and the absensi
// anchor. It carries no soft-delete columns: a session is cancelled
// (status=dibatalkan + alasan_batal), never deleted.
type Sesi struct {
	ID            int64      `gorm:"column:id;primaryKey"          json:"id"`
	JadwalID      int64      `gorm:"column:jadwal_id"              json:"jadwal_id"`
	TanggalLokal  time.Time  `gorm:"column:tanggal_lokal;type:date" json:"tanggal_lokal"`
	MulaiUTC      time.Time  `gorm:"column:mulai_utc"              json:"mulai_utc"`
	SelesaiUTC    time.Time  `gorm:"column:selesai_utc"            json:"selesai_utc"`
	Timezone      string     `gorm:"column:timezone"               json:"timezone"`
	AbsenBukaUTC  time.Time  `gorm:"column:absen_buka_utc"         json:"absen_buka_utc"`
	AbsenTutupUTC time.Time  `gorm:"column:absen_tutup_utc"        json:"absen_tutup_utc"`
	Status        StatusSesi `gorm:"column:status"                 json:"status"`
	AlasanBatal   *string    `gorm:"column:alasan_batal"           json:"alasan_batal,omitempty"`
	JmlDitugaskan int        `gorm:"column:jml_ditugaskan"         json:"jml_ditugaskan"`
	JmlHadir      int        `gorm:"column:jml_hadir"              json:"jml_hadir"`
	Catatan       *string    `gorm:"column:catatan"                json:"catatan,omitempty"`
	CreatedAt     time.Time  `gorm:"column:created_at"             json:"created_at"`
	CreatedBy     *int64     `gorm:"column:created_by"             json:"created_by,omitempty"`
	ModifiedAt    *time.Time `gorm:"column:modified_at"            json:"modified_at,omitempty"`
	ModifiedBy    *int64     `gorm:"column:modified_by"            json:"modified_by,omitempty"`
}

// TableName pins the table name for GORM.
func (Sesi) TableName() string { return "jadwal_sesi" }
