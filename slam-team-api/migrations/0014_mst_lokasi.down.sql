-- Rollback: 0014 mst_lokasi

ALTER TABLE mst_lokasi DROP CONSTRAINT IF EXISTS fk_lokasi_deleted_by;
ALTER TABLE mst_lokasi DROP CONSTRAINT IF EXISTS fk_lokasi_modified_by;
ALTER TABLE mst_lokasi DROP CONSTRAINT IF EXISTS fk_lokasi_created_by;
ALTER TABLE mst_lokasi DROP CONSTRAINT IF EXISTS fk_lokasi_foto;

DROP INDEX IF EXISTS ix_mst_lokasi_is_aktif;
DROP INDEX IF EXISTS ux_mst_lokasi_kode;

DROP TABLE IF EXISTS mst_lokasi;
