-- Rollback: 0005 mst_instansi

ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_deleted_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_modified_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_created_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_logo_tambahan;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_logo_utama;

DROP INDEX IF EXISTS ix_mst_instansi_status;
DROP INDEX IF EXISTS ux_mst_instansi_kode;

DROP TABLE IF EXISTS mst_instansi;
