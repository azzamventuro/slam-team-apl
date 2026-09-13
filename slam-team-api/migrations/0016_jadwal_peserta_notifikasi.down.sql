-- Rollback: 0016 jadwal_peserta + notifikasi
-- Both tables are leaves (nothing references them). Enum types stay — they
-- belong to 0001.

ALTER TABLE notifikasi DROP CONSTRAINT IF EXISTS fk_notifikasi_created_by;
ALTER TABLE notifikasi DROP CONSTRAINT IF EXISTS fk_notifikasi_user;

DROP INDEX IF EXISTS ix_notifikasi_kedaluwarsa;
DROP INDEX IF EXISTS ix_notifikasi_reff;
DROP INDEX IF EXISTS ix_notifikasi_user_dibaca_created;

DROP TABLE IF EXISTS notifikasi;

ALTER TABLE jadwal_peserta DROP CONSTRAINT IF EXISTS fk_jadwal_peserta_deleted_by;
ALTER TABLE jadwal_peserta DROP CONSTRAINT IF EXISTS fk_jadwal_peserta_ditugaskan_oleh;
ALTER TABLE jadwal_peserta DROP CONSTRAINT IF EXISTS fk_jadwal_peserta_anggota;
ALTER TABLE jadwal_peserta DROP CONSTRAINT IF EXISTS fk_jadwal_peserta_sesi;
ALTER TABLE jadwal_peserta DROP CONSTRAINT IF EXISTS fk_jadwal_peserta_jadwal;

DROP INDEX IF EXISTS ix_jadwal_peserta_jadwal_sesi;
DROP INDEX IF EXISTS ix_jadwal_peserta_anggota_jadwal;

DROP TABLE IF EXISTS jadwal_peserta;
