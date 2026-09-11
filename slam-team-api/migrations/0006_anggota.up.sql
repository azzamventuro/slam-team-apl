-- Migration 0006: mst_wilayah + anggota — Master Data Anggota (core member record)
-- Created: 2026-09-09
-- Prerequisites: 0001_enums (jenis_anggota, status_anggota), 0003_file_layer (mst_file),
--                0005_mst_instansi (mst_instansi).

-- ─── mst_wilayah ────────────────────────────────────────────────────────────
-- Lightweight reference table (supplies 4-digit NRA prefix). Created here
-- because anggota.wilayah_id is an FK to it.

CREATE TABLE IF NOT EXISTS mst_wilayah (
    id         bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode       varchar(10)  NOT NULL,
    nama       varchar(150) NOT NULL,
    provinsi   varchar(100),
    tingkat    varchar(20),
    is_aktif   boolean      NOT NULL DEFAULT true,
    created_at timestamptz  NOT NULL DEFAULT now(),
    created_by bigint
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_mst_wilayah_kode ON mst_wilayah (kode);

-- ─── anggota ────────────────────────────────────────────────────────────────

CREATE TABLE anggota (
    id                      bigint           GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    instansi_id             bigint           NOT NULL,
    wilayah_id              bigint,
    no_induk                varchar(20),
    nama_lengkap            varchar(150)     NOT NULL,
    nama_panggilan          varchar(50),
    foto_profil_file_id     bigint,
    foto_formal_file_id     bigint,
    jenis_anggota           jenis_anggota    NOT NULL,
    jenis_kelamin           int,
    jenis_identitas         int,
    no_identitas            varchar(50),
    file_identitas_file_id  bigint,
    pekerjaan               varchar(100),
    alamat                  text,
    kode_pos                varchar(10),
    tempat_lahir            varchar(100),
    tanggal_lahir           date             NOT NULL,
    tanggal_bergabung       date,
    status_anggota          status_anggota   NOT NULL DEFAULT 'aktif',
    is_deleted              boolean          NOT NULL DEFAULT false,
    deleted_at              timestamptz,
    deleted_by              bigint,
    created_at              timestamptz      NOT NULL DEFAULT now(),
    created_by              bigint,
    modified_at             timestamptz,
    modified_by             bigint
);

-- Indexes (from DBML).
CREATE UNIQUE INDEX ux_anggota_no_induk ON anggota (no_induk) WHERE is_deleted = false;
CREATE INDEX ix_anggota_instansi_id ON anggota (instansi_id);
CREATE INDEX ix_anggota_jenis_anggota ON anggota (jenis_anggota);
CREATE INDEX ix_anggota_status_anggota ON anggota (status_anggota);

-- ─── Deferred FK from 0005: users.anggota_id → anggota.id ────────────────
-- Now that anggota exists, add the FK that was deferred from the users table.
ALTER TABLE users ADD CONSTRAINT fk_users_anggota
    FOREIGN KEY (anggota_id) REFERENCES anggota (id) ON DELETE RESTRICT;

-- FK: instansi (restrict — cannot delete instansi referenced by anggota).
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_instansi
    FOREIGN KEY (instansi_id) REFERENCES mst_instansi (id) ON DELETE RESTRICT;

-- FK: wilayah (nullable, no restrict needed — wilayah is reference-only).
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_wilayah
    FOREIGN KEY (wilayah_id) REFERENCES mst_wilayah (id) ON DELETE SET NULL;

-- FK: file references (nullable).
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_foto_profil
    FOREIGN KEY (foto_profil_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_foto_formal
    FOREIGN KEY (foto_formal_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_file_identitas
    FOREIGN KEY (file_identitas_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;

-- FK: audit / soft-delete user references (nullable).
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE anggota ADD CONSTRAINT fk_anggota_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;

-- ─── Permission seed (idempotent) ──────────────────────────────────────────
-- mst_permission columns: id, modul_id (FK→mst_modul), aksi (enum), kode, nama, keterangan
-- The 'anggota' module row in mst_modul must exist. Seed it idempotently.

INSERT INTO mst_modul (kode, nama, grup, urutan, tampil_di_menu) VALUES
    ('anggota', 'Anggota', 'Master Data', 6, true)
ON CONFLICT (kode) DO NOTHING;

INSERT INTO mst_permission (modul_id, aksi, kode, nama, keterangan) VALUES
    ((SELECT id FROM mst_modul WHERE kode = 'anggota'), 'create', 'anggota.create', 'Buat Anggota', 'Membuat anggota baru'),
    ((SELECT id FROM mst_modul WHERE kode = 'anggota'), 'read',   'anggota.read',   'Lihat Anggota', 'Melihat data anggota'),
    ((SELECT id FROM mst_modul WHERE kode = 'anggota'), 'update', 'anggota.update', 'Ubah Anggota', 'Mengubah data anggota'),
    ((SELECT id FROM mst_modul WHERE kode = 'anggota'), 'delete', 'anggota.delete', 'Hapus Anggota', 'Menghapus (soft delete) anggota')
ON CONFLICT (kode) DO NOTHING;

-- Seed default role_permissions using role kode lookups (not hardcoded IDs).
-- Admin (kode='admin') = full CRUD, cakupan semua.
INSERT INTO role_permission (role_id, permission_id, cakupan)
SELECT r.id, p.id, 'semua'
FROM mst_role r, mst_permission p
WHERE r.kode = 'admin' AND p.kode IN ('anggota.create','anggota.read','anggota.update','anggota.delete')
ON CONFLICT DO NOTHING;

-- Moderator (kode='moderator') = full CRUD, cakupan semua.
INSERT INTO role_permission (role_id, permission_id, cakupan)
SELECT r.id, p.id, 'semua'
FROM mst_role r, mst_permission p
WHERE r.kode = 'moderator' AND p.kode IN ('anggota.create','anggota.read','anggota.update','anggota.delete')
ON CONFLICT DO NOTHING;

-- User (kode='user') = read only, cakupan semua.
INSERT INTO role_permission (role_id, permission_id, cakupan)
SELECT r.id, p.id, 'semua'
FROM mst_role r, mst_permission p
WHERE r.kode = 'user' AND p.kode = 'anggota.read'
ON CONFLICT DO NOTHING;
