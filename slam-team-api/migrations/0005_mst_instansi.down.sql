-- Rollback: 0005 users + mst_instansi

-- Drop mst_instansi first (depends on users).
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_deleted_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_modified_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_created_by;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_logo_tambahan;
ALTER TABLE mst_instansi DROP CONSTRAINT IF EXISTS fk_instansi_logo_utama;

DROP INDEX IF EXISTS ix_mst_instansi_status;
DROP INDEX IF EXISTS ux_mst_instansi_kode;

DROP TABLE IF EXISTS mst_instansi;

-- Drop users (no longer depends on anything in this migration).
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_created_by;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_modified_by;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_deleted_by;
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_role;

DROP INDEX IF EXISTS ix_users_role_id;
DROP INDEX IF EXISTS ix_users_anggota_id;
DROP INDEX IF EXISTS ux_users_email;
DROP INDEX IF EXISTS ux_users_username;

DROP TABLE IF EXISTS users;

-- Rollback deferred FKs from 0004_rbac that were added in this migration.
ALTER TABLE user_role DROP CONSTRAINT IF EXISTS fk_ur_created_by;
ALTER TABLE user_role DROP CONSTRAINT IF EXISTS fk_ur_user;
ALTER TABLE role_permission DROP CONSTRAINT IF EXISTS fk_rp_created_by;
ALTER TABLE mst_role DROP CONSTRAINT IF EXISTS fk_role_created_by;
ALTER TABLE mst_role DROP CONSTRAINT IF EXISTS fk_role_modified_by;
ALTER TABLE mst_role DROP CONSTRAINT IF EXISTS fk_role_deleted_by;
ALTER TABLE mst_modul DROP CONSTRAINT IF EXISTS fk_modul_created_by;
ALTER TABLE mst_modul DROP CONSTRAINT IF EXISTS fk_modul_modified_by;
ALTER TABLE mst_pengaturan DROP CONSTRAINT IF EXISTS fk_pengaturan_modified_by;
