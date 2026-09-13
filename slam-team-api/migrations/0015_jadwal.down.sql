-- Rollback: 0015 jadwal + jadwal_sesi
-- Children first (jadwal_sesi → jadwal). Enum types stay — they belong to 0001.

ALTER TABLE jadwal_sesi DROP CONSTRAINT IF EXISTS fk_jadwal_sesi_modified_by;
ALTER TABLE jadwal_sesi DROP CONSTRAINT IF EXISTS fk_jadwal_sesi_created_by;
ALTER TABLE jadwal_sesi DROP CONSTRAINT IF EXISTS fk_jadwal_sesi_jadwal;

DROP INDEX IF EXISTS ix_jadwal_sesi_jendela;
DROP INDEX IF EXISTS ix_jadwal_sesi_tanggal_status;

DROP TABLE IF EXISTS jadwal_sesi;

ALTER TABLE jadwal DROP CONSTRAINT IF EXISTS fk_jadwal_deleted_by;
ALTER TABLE jadwal DROP CONSTRAINT IF EXISTS fk_jadwal_modified_by;
ALTER TABLE jadwal DROP CONSTRAINT IF EXISTS fk_jadwal_created_by;
ALTER TABLE jadwal DROP CONSTRAINT IF EXISTS fk_jadwal_inorga;
ALTER TABLE jadwal DROP CONSTRAINT IF EXISTS fk_jadwal_lokasi;

DROP INDEX IF EXISTS ix_jadwal_lokasi_id;
DROP INDEX IF EXISTS ix_jadwal_tanggal_mulai_status;
DROP INDEX IF EXISTS ux_jadwal_kode;

DROP TABLE IF EXISTS jadwal;
