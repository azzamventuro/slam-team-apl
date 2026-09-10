-- 0004_rbac — "Kelompok 1 — Hak Akses" from slamteam_db.dbml: 5 tables
-- that power the dynamic RBAC (matriks izin sebagai data, bukan kode).
--
-- ENUMS: aksi_permission and cakupan_permission are created by 0001_enums.
-- ORDER: after 0003_file_layer (which creates enums); before any table that
-- uses mst_modul FK (none yet — seed comes via `slamctl seed rbac`).
--
-- FOREIGN KEYS: created_by / modified_by / deleted_by → users.id are deferred
-- to a later migration (users table doesn't exist yet). role_permission and
-- user_role FKs to mst_role are immediate because mst_role is in this file.
-- user_role FK to users.id is deferred (same reason).

-- =====================================================================
--  mst_modul — 19 menu modules + 1 non-menu (file)
-- =====================================================================
CREATE TABLE mst_modul (
    id              bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode            varchar(50)  NOT NULL,
    nama            varchar(100) NOT NULL,
    grup            varchar(50),
    route           varchar(150),
    icon            varchar(50),
    urutan          int          NOT NULL DEFAULT 0,
    tampil_di_menu  boolean      NOT NULL DEFAULT true,
    is_aktif        boolean      NOT NULL DEFAULT true,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    created_by      bigint,
    modified_at     timestamptz,
    modified_by     bigint,

    CONSTRAINT uq_modul_kode UNIQUE (kode)
);

-- =====================================================================
--  mst_permission — modul × aksi pairs (the permission catalogue)
-- =====================================================================
CREATE TABLE mst_permission (
    id            bigint          GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    modul_id      bigint          NOT NULL,
    aksi          aksi_permission NOT NULL,
    kode          varchar(100)    NOT NULL,
    nama          varchar(150),
    keterangan    text,
    is_berbahaya  boolean         NOT NULL DEFAULT false,

    CONSTRAINT uq_permission_kode UNIQUE (kode),
    CONSTRAINT uq_permission_modul_aksi UNIQUE (modul_id, aksi),
    CONSTRAINT fk_permission_modul FOREIGN KEY (modul_id)
        REFERENCES mst_modul (id) ON DELETE RESTRICT
);

-- =====================================================================
--  mst_role — roles with level hierarchy + super admin bypass
-- =====================================================================
CREATE TABLE mst_role (
    id           bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode         varchar(50)  NOT NULL,
    nama         varchar(100) NOT NULL,
    level        int          NOT NULL,
    keterangan   text,
    is_sistem    boolean      NOT NULL DEFAULT false,
    is_super     boolean      NOT NULL DEFAULT false,
    is_aktif     boolean      NOT NULL DEFAULT true,
    is_deleted   boolean      NOT NULL DEFAULT false,
    deleted_at   timestamptz,
    deleted_by   bigint,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    created_by   bigint,
    modified_at  timestamptz,
    modified_by  bigint,

    CONSTRAINT uq_role_kode UNIQUE (kode)
);

CREATE INDEX ix_role_active ON mst_role (is_deleted, level);

-- =====================================================================
--  role_permission — the matrix (Bab 3.5). Heart of the RBAC.
-- =====================================================================
CREATE TABLE role_permission (
    id             bigint             GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    role_id        bigint             NOT NULL,
    permission_id  bigint             NOT NULL,
    cakupan        cakupan_permission NOT NULL DEFAULT 'semua',
    created_at     timestamptz        NOT NULL DEFAULT now(),
    created_by     bigint,

    CONSTRAINT uq_role_permission UNIQUE (role_id, permission_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id)
        REFERENCES mst_role (id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_permission FOREIGN KEY (permission_id)
        REFERENCES mst_permission (id) ON DELETE RESTRICT
);

CREATE INDEX ix_rp_role ON role_permission (role_id);

-- =====================================================================
--  user_role — user ↔ role mapping (supports future multi-role)
-- =====================================================================
CREATE TABLE user_role (
    id              bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         bigint       NOT NULL,
    role_id         bigint       NOT NULL,
    is_utama        boolean      NOT NULL DEFAULT true,
    berlaku_sampai  timestamptz,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    created_by      bigint,

    CONSTRAINT uq_user_role UNIQUE (user_id, role_id),
    CONSTRAINT fk_ur_role FOREIGN KEY (role_id)
        REFERENCES mst_role (id) ON DELETE RESTRICT
    -- FK to users.id deferred to users migration
);

CREATE INDEX ix_ur_user ON user_role (user_id);
CREATE INDEX ix_ur_role ON user_role (role_id);
