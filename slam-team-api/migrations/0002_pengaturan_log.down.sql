-- Rollback of 0002_pengaturan_log. Indexes are dropped with their table; the
-- tipe_nilai_pengaturan enum belongs to 0001_enums and rolls back with it.

DROP TABLE IF EXISTS log_aktivitas;
DROP TABLE IF EXISTS mst_pengaturan;
