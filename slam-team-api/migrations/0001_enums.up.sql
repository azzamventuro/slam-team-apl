-- 0001_enums — every Postgres ENUM declared in slamteam_db.dbml.
--
-- Enums are created once, before any table that references them. Adding a new
-- value later is a NEW migration (ALTER TYPE ... ADD VALUE), never an edit to
-- this file.

CREATE TYPE aksi_permission AS ENUM (
    'create',      -- also: melakukan absensi, mengajukan izin, mengunggah berkas
    'read',
    'update',
    'delete',
    'approve',
    'assign',
    'export',
    'override',    -- membetulkan absensi yang gagal atau terlewat
    'print',
    'cabut',       -- mencabut KTA; barisnya tetap disimpan, jadi bukan delete
    'batal_sesi'   -- membatalkan satu sesi jadwal (berdampak ke semua peserta)
);

CREATE TYPE cakupan_permission AS ENUM (
    'semua',
    'instansi_sendiri',
    'milik_sendiri'
);

CREATE TYPE jenis_anggota AS ENUM (
    'siswa',
    'dewasa',
    'siswa_ke_dewasa'  -- wajib punya KTA pelajar DAN KTA dewasa
);

CREATE TYPE status_anggota AS ENUM (
    'aktif',      -- diset manual; KTA kadaluarsa TIDAK menonaktifkan anggota
    'non_aktif'
);

CREATE TYPE jenis_kta AS ENUM (
    'pelajar',
    'dewasa'
);

CREATE TYPE status_kta AS ENUM (
    'aktif',
    'arsip',       -- KTA pelajar yang sudah digantikan KTA dewasa
    'dicabut',
    'kadaluarsa'
);

CREATE TYPE mode_absen AS ENUM (
    'masuk_saja',
    'masuk_pulang'
);

CREATE TYPE tipe_absensi AS ENUM (
    'masuk',
    'pulang'
);

CREATE TYPE metode_absensi AS ENUM (
    'selfie',   -- absen mandiri lewat aplikasi (jalur normal)
    'qr',
    'manual'    -- dibuat moderator lewat override
);

CREATE TYPE status_kehadiran AS ENUM (
    'hadir',
    'terlambat',
    'pulang_cepat',
    'hadir_luar_radius',
    'izin',
    'sakit',
    'dinas',
    'alfa'
);

CREATE TYPE jenis_izin AS ENUM (
    'izin',
    'sakit',
    'dinas',
    'pulang_cepat'
);

CREATE TYPE status_izin AS ENUM (
    'menunggu',
    'disetujui',
    'ditolak'
);

CREATE TYPE status_sesi AS ENUM (
    'terjadwal',
    'berlangsung',
    'selesai',
    'dibatalkan'
);

CREATE TYPE status_tugas AS ENUM (
    'ditugaskan',
    'diterima',
    'ditolak',
    'izin'
);

CREATE TYPE pola_ulang AS ENUM (
    'tidak_berulang',
    'harian',
    'mingguan',
    'bulanan',
    'kustom'
);

CREATE TYPE varian_file AS ENUM (
    'original',
    'medium',
    'low'
);

CREATE TYPE status_proses_file AS ENUM (
    'menunggu',
    'selesai',
    'gagal'
);

CREATE TYPE tipe_nilai_pengaturan AS ENUM (
    'string',
    'integer',
    'boolean',
    'json',
    'date'
);

CREATE TYPE prioritas_notifikasi AS ENUM (
    'rendah',
    'normal',
    'tinggi'
);
