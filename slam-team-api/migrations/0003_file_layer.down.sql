-- Rollback of 0003_file_layer. mst_file_varian goes first: its FK to mst_file
-- is ON DELETE CASCADE for rows, not for the table itself. Indexes and the
-- unique constraint are dropped with their tables; the varian_file and
-- status_proses_file enums belong to 0001_enums and roll back with it.
--
-- Bytes on disk under STORAGE_ROOT are NOT touched — a schema rollback must
-- not destroy uploaded evidence.

DROP TABLE IF EXISTS mst_file_varian;
DROP TABLE IF EXISTS mst_file;
