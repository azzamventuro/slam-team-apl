-- Migration 0006 rollback: mst_wilayah + anggota

-- Remove seeded permissions (role_permission first due to FK).
DELETE FROM role_permission WHERE permission_id IN (
    SELECT id FROM mst_permission WHERE modul = 'anggota'
);
DELETE FROM mst_permission WHERE modul = 'anggota';

-- Drop deferred FK from 0005: users.anggota_id → anggota.id
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_anggota;

-- Drop anggota table (cascades FKs referencing it: users, kta, jadwal_peserta, etc.
-- are not created yet, so no data-loss risk).
DROP TABLE IF EXISTS anggota;

-- Drop mst_wilayah (only created in this migration).
DROP TABLE IF EXISTS mst_wilayah;
