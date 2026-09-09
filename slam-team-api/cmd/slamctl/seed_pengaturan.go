package main

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// init wires the pengaturan seeder into `slamctl seed`. Reference data belongs
// to the module that owns its table, so the registration lives beside the rows
// rather than in the generic runner.
func init() {
	seedOrder = append(seedOrder, "pengaturan")
	seeders["pengaturan"] = seedPengaturan
}

// pengaturanRow is one seeded setting. nilai is the initial value AND the
// nilai_bawaan the "restore default" button resets to — they start identical
// and only nilai ever diverges.
type pengaturanRow struct {
	kunci      string
	grup       string
	label      string
	nilai      string
	tipe       string
	opsi       *string // JSON array, or nil when the control is not a dropdown
	satuan     string
	keterangan string
	urutan     int
	publik     bool
	terkunci   bool
}

// opsiOf makes a JSON array literal for the opsi column.
func opsiOf(s string) *string { return &s }

// seedPengaturanRows is the SEED AWAL list from slamteam_db.dbml. Every
// threshold the system reads at runtime is here — variant pixel sizes, KTA
// dimensions and validity, absen radius and windows, the NRA region code — so
// none of them is hardcoded in Go or Angular.
//
// is_publik marks what the unauthenticated app shell may read (identity only,
// never a threshold). is_terkunci marks what only a super admin may change:
// values that decide how records are numbered or how strictly attendance is
// judged, where a careless edit silently rewrites the rules.
var seedPengaturanRows = []pengaturanRow{
	// --- umum ---
	{kunci: "umum.nama_klub", grup: "umum", label: "Nama klub",
		nilai: "Scouting Legion Airsofter Malang", tipe: "string", urutan: 1, publik: true,
		keterangan: "Tampil di header aplikasi dan tercetak di KTA"},
	{kunci: "umum.singkatan", grup: "umum", label: "Singkatan",
		nilai: "SLAM", tipe: "string", urutan: 2, publik: true},
	{kunci: "umum.alamat_sekretariat", grup: "umum", label: "Alamat sekretariat",
		nilai: "Jalan Laksda Adi Sucipto Nomor 32 Blimbing, Kota Malang", tipe: "string", urutan: 3, publik: true,
		keterangan: "Sumber tunggal alamat yang tercetak di balik KTA"},
	{kunci: "umum.timezone_bawaan", grup: "umum", label: "Zona waktu bawaan",
		nilai: "Asia/Jakarta", tipe: "string", urutan: 4, publik: true,
		keterangan: "Zona IANA untuk user yang belum memilih zona sendiri"},
	{kunci: "umum.kode_wilayah_paten", grup: "umum", label: "Kode wilayah untuk NRA",
		nilai: "3573", tipe: "string", urutan: 5, terkunci: true,
		keterangan: "Awalan NRA; mengubahnya memutus deret nomor yang sudah terbit"},

	// --- nra ---
	{kunci: "nra.mode_penomoran", grup: "nra", label: "Mode penomoran NRA",
		nilai: "global", tipe: "string", opsi: opsiOf(`["global","per_anggota"]`),
		urutan: 10, terkunci: true},
	{kunci: "nra.panjang_urut_minimal", grup: "nra", label: "Panjang minimal nomor urut",
		nilai: "3", tipe: "integer", satuan: "digit", urutan: 11, terkunci: true,
		keterangan: "Melar sendiri ke 4 digit setelah 999"},

	// --- kta ---
	{kunci: "kta.masa_berlaku_pelajar_bulan", grup: "kta", label: "Masa berlaku KTA pelajar",
		nilai: "12", tipe: "integer", satuan: "bulan", urutan: 20,
		keterangan: "Dihitung dari tanggal terbit kartu"},
	{kunci: "kta.masa_berlaku_dewasa_bulan", grup: "kta", label: "Masa berlaku KTA dewasa",
		nilai: "24", tipe: "integer", satuan: "bulan", urutan: 21,
		keterangan: "Dihitung dari tanggal terbit kartu"},
	{kunci: "kta.format_keluaran", grup: "kta", label: "Format berkas kartu",
		nilai: "png", tipe: "string", opsi: opsiOf(`["png"]`), urutan: 22},
	{kunci: "kta.dpi_cetak", grup: "kta", label: "Resolusi cetak",
		nilai: "300", tipe: "integer", satuan: "dpi", urutan: 23},
	{kunci: "kta.lebar_mm", grup: "kta", label: "Lebar kartu",
		nilai: "85.6", tipe: "string", satuan: "mm", urutan: 24,
		keterangan: "Standar ID-1 / CR80; pecahan, jadi disimpan sebagai teks"},
	{kunci: "kta.tinggi_mm", grup: "kta", label: "Tinggi kartu",
		nilai: "54", tipe: "integer", satuan: "mm", urutan: 25,
		keterangan: "Standar ID-1 / CR80"},
	{kunci: "kta.bleed_mm", grup: "kta", label: "Bleed setiap sisi",
		nilai: "2", tipe: "integer", satuan: "mm", urutan: 26},
	{kunci: "kta.qr_ukuran_mm", grup: "kta", label: "Ukuran kotak QR",
		nilai: "24", tipe: "integer", satuan: "mm", urutan: 27},
	{kunci: "kta.kartu_per_lembar", grup: "kta", label: "Kartu per lembar cetak massal",
		nilai: "8", tipe: "integer", urutan: 28},

	// --- absensi ---
	{kunci: "absensi.akurasi_gps_maks_meter", grup: "absensi", label: "Akurasi GPS maksimum",
		nilai: "100", tipe: "integer", satuan: "meter", urutan: 30, terkunci: true,
		keterangan: "Absensi ditolak bila akurasi perangkat lebih buruk dari ini"},
	{kunci: "absensi.radius_bawaan_meter", grup: "absensi", label: "Radius bawaan lokasi",
		nilai: "100", tipe: "integer", satuan: "meter", urutan: 31,
		keterangan: "Dipakai saat lokasi baru dibuat tanpa radius sendiri"},
	{kunci: "absensi.toleransi_telat_menit", grup: "absensi", label: "Toleransi keterlambatan",
		nilai: "15", tipe: "integer", satuan: "menit", urutan: 32},
	{kunci: "absensi.buka_absen_menit", grup: "absensi", label: "Jendela absen dibuka sebelum",
		nilai: "30", tipe: "integer", satuan: "menit", urutan: 33},
	{kunci: "absensi.tutup_absen_menit", grup: "absensi", label: "Jendela absen ditutup setelah",
		nilai: "60", tipe: "integer", satuan: "menit", urutan: 34},

	// --- file ---
	{kunci: "file.varian_original_maks_px", grup: "file", label: "Sisi terpanjang varian original",
		nilai: "4000", tipe: "integer", satuan: "px", urutan: 40, terkunci: true,
		keterangan: "Unggahan yang lebih besar dikecilkan sampai sisi ini"},
	{kunci: "file.varian_medium_px", grup: "file", label: "Sisi terpanjang varian medium",
		nilai: "1200", tipe: "integer", satuan: "px", urutan: 41, terkunci: true},
	{kunci: "file.varian_low_px", grup: "file", label: "Sisi terpanjang varian low",
		nilai: "400", tipe: "integer", satuan: "px", urutan: 42, terkunci: true},
}

// seedPengaturan is idempotent: it is keyed on the natural unique column
// (kunci), so a rerun refreshes the metadata a new release changed — label,
// keterangan, opsi, satuan, urutan, flags, nilai_bawaan — and NEVER touches
// nilai. The live value belongs to the operator; re-seeding must not undo an
// edit they made in the settings page.
func seedPengaturan(ctx context.Context, db *gorm.DB) error {
	ok, err := tableExists(db, "mst_pengaturan")
	if err != nil {
		return fmt.Errorf("periksa tabel mst_pengaturan: %w", err)
	}
	if !ok {
		return fmt.Errorf("tabel mst_pengaturan belum ada — jalankan `slamctl migrate up` dulu")
	}

	const upsert = `
		INSERT INTO mst_pengaturan
			(kunci, grup, label, nilai, tipe_nilai, nilai_bawaan, opsi,
			 satuan, keterangan, urutan, is_publik, is_terkunci)
		VALUES (?, ?, ?, ?, CAST(? AS tipe_nilai_pengaturan), ?, CAST(? AS jsonb),
		        NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?)
		ON CONFLICT (kunci) DO UPDATE SET
			grup         = EXCLUDED.grup,
			label        = EXCLUDED.label,
			tipe_nilai   = EXCLUDED.tipe_nilai,
			nilai_bawaan = EXCLUDED.nilai_bawaan,
			opsi         = EXCLUDED.opsi,
			satuan       = EXCLUDED.satuan,
			keterangan   = EXCLUDED.keterangan,
			urutan       = EXCLUDED.urutan,
			is_publik    = EXCLUDED.is_publik,
			is_terkunci  = EXCLUDED.is_terkunci`

	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, r := range seedPengaturanRows {
			if err := tx.Exec(upsert,
				r.kunci, r.grup, r.label, r.nilai, r.tipe, r.nilai, r.opsi,
				r.satuan, r.keterangan, r.urutan, r.publik, r.terkunci,
			).Error; err != nil {
				return fmt.Errorf("upsert %q: %w", r.kunci, err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	fmt.Printf("(%d setelan) ", len(seedPengaturanRows))
	return nil
}
