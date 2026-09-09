-- Rollback of 0001_enums. Dropped in reverse order of creation; every table
-- using them belongs to a later migration, which rolls back first.

DROP TYPE IF EXISTS prioritas_notifikasi;
DROP TYPE IF EXISTS tipe_nilai_pengaturan;
DROP TYPE IF EXISTS status_proses_file;
DROP TYPE IF EXISTS varian_file;
DROP TYPE IF EXISTS pola_ulang;
DROP TYPE IF EXISTS status_tugas;
DROP TYPE IF EXISTS status_sesi;
DROP TYPE IF EXISTS status_izin;
DROP TYPE IF EXISTS jenis_izin;
DROP TYPE IF EXISTS status_kehadiran;
DROP TYPE IF EXISTS metode_absensi;
DROP TYPE IF EXISTS tipe_absensi;
DROP TYPE IF EXISTS mode_absen;
DROP TYPE IF EXISTS status_kta;
DROP TYPE IF EXISTS jenis_kta;
DROP TYPE IF EXISTS status_anggota;
DROP TYPE IF EXISTS jenis_anggota;
DROP TYPE IF EXISTS cakupan_permission;
DROP TYPE IF EXISTS aksi_permission;
