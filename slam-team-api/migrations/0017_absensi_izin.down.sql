-- Rollback: 0017 absensi_izin
-- The table is a leaf until the absensi migration adds absensi.izin_id (that
-- migration drops its own FK first). Enum types stay — they belong to 0001.

ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_deleted_by;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_modified_by;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_created_by;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_diproses_oleh;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_lampiran;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_anggota;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_jadwal;
ALTER TABLE absensi_izin DROP CONSTRAINT IF EXISTS fk_absensi_izin_sesi;

DROP INDEX IF EXISTS ix_absensi_izin_status_created;
DROP INDEX IF EXISTS ix_absensi_izin_sesi_anggota;

DROP TABLE IF EXISTS absensi_izin;
