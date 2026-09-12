-- Migration 0005: users + mst_instansi
-- Created: 2026-09-10
--
-- Creates the users table first (needed by mst_instansi audit FKs).
-- The FK users.anggota_id → anggota.id is DEFERRED to 0006_anggota
-- because the anggota table doesn't exist yet.

-- ─── users ────────────────────────────────────────────────────────────────
CREATE TABLE users (
    id                bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    anggota_id        bigint,
    username          varchar(50)   NOT NULL,
    email             varchar(150)  NOT NULL,
    password          varchar(255),
    role_id           bigint        NOT NULL,
    timezone          varchar(50)   NOT NULL DEFAULT 'Asia/Jakarta',
    is_aktif          boolean       NOT NULL DEFAULT true,
    login_terakhir    timestamptz,
    password_diubah   timestamptz,
    gagal_login       int           NOT NULL DEFAULT 0,
    terkunci_sampai   timestamptz,
    is_deleted        boolean       NOT NULL DEFAULT false,
    deleted_at        timestamptz,
    deleted_by        bigint,
    created_at        timestamptz   NOT NULL DEFAULT now(),
    created_by        bigint,
    modified_at       timestamptz,
    modified_by       bigint
);

CREATE UNIQUE INDEX ux_users_username ON users (username) WHERE is_deleted = false;
CREATE UNIQUE INDEX ux_users_email    ON users (email)    WHERE is_deleted = false;
CREATE INDEX ix_users_anggota_id      ON users (anggota_id);
CREATE INDEX ix_users_role_id         ON users (role_id);

-- FK: role (immediate — mst_role exists from 0004).
ALTER TABLE users ADD CONSTRAINT fk_users_role
    FOREIGN KEY (role_id) REFERENCES mst_role (id) ON DELETE RESTRICT;

-- FK: audit / soft-delete user self-references (nullable).
ALTER TABLE users ADD CONSTRAINT fk_users_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE users ADD CONSTRAINT fk_users_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE users ADD CONSTRAINT fk_users_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;

-- FK to mst_instansi is NOT here because instansi doesn't exist yet.
-- FK to anggota is NOT here because anggota doesn't exist yet.

-- ─── Deferred FKs from 0004_rbac ─────────────────────────────────────────
-- user_role.user_id → users.id (deferred in 0004 because users didn't exist).
ALTER TABLE user_role ADD CONSTRAINT fk_ur_user
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- Audit FKs on 0004 tables → users.id (deferred in 0004).
ALTER TABLE mst_modul ADD CONSTRAINT fk_modul_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_modul ADD CONSTRAINT fk_modul_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE mst_role ADD CONSTRAINT fk_role_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_role ADD CONSTRAINT fk_role_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_role ADD CONSTRAINT fk_role_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE role_permission ADD CONSTRAINT fk_rp_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;

ALTER TABLE user_role ADD CONSTRAINT fk_ur_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;

-- Audit FK on 0002_pengaturan → users.id (deferred).
ALTER TABLE mst_pengaturan ADD CONSTRAINT fk_pengaturan_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;

-- ─── mst_instansi ───────────────────────────────────────────────────────
CREATE TABLE mst_instansi (
    id                     bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode                   varchar(50)   NOT NULL,
    nama                   varchar(150)  NOT NULL,
    nama_club              varchar(150)  NOT NULL DEFAULT '',
    alamat                 text          NOT NULL DEFAULT '',
    no_telepon             varchar(30)   NOT NULL DEFAULT '',
    logo_utama_file_id     bigint,
    logo_tambahan_file_id  bigint,
    tanggal_bergabung      date,
    status                 int           NOT NULL DEFAULT 1,
    is_deleted             boolean       NOT NULL DEFAULT false,
    deleted_at             timestamptz,
    deleted_by             bigint,
    created_at             timestamptz   NOT NULL DEFAULT now(),
    created_by             bigint,
    modified_at            timestamptz,
    modified_by            bigint
);

-- Partial-unique index: kode must be unique among non-deleted rows.
CREATE UNIQUE INDEX ux_mst_instansi_kode ON mst_instansi (kode) WHERE is_deleted = false;

-- Filter by status (aktif / nonaktif).
CREATE INDEX ix_mst_instansi_status ON mst_instansi (status);

-- FK: logo file references (nullable).
ALTER TABLE mst_instansi ADD CONSTRAINT fk_instansi_logo_utama
    FOREIGN KEY (logo_utama_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;
ALTER TABLE mst_instansi ADD CONSTRAINT fk_instansi_logo_tambahan
    FOREIGN KEY (logo_tambahan_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;

-- FK: audit / soft-delete user references (nullable).
ALTER TABLE mst_instansi ADD CONSTRAINT fk_instansi_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_instansi ADD CONSTRAINT fk_instansi_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_instansi ADD CONSTRAINT fk_instansi_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;
