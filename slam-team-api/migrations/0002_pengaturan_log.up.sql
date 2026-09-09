-- 0002_pengaturan_log — the two "Kelompok 2 — Sistem" tables from
-- slamteam_db.dbml: typed settings and the generic audit trail.
--
-- Both land before every feature module because almost all of them read a
-- threshold from mst_pengaturan (variant px, KTA validity, absen radius) and
-- write a log_aktivitas row on state changes.
--
-- FOREIGN KEYS: .dbml declares mst_pengaturan.modified_by > users.id and
-- log_aktivitas.aktor_user_id > users.id. The users table belongs to a LATER
-- migration (auth/identity), so the columns are created plain here and the two
-- constraints are added by that migration — a FK cannot precede its target.

CREATE TABLE mst_pengaturan (
    id           bigint                GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- Natural key, format "grup.nama" (e.g. "kta.masa_berlaku_pelajar_bulan").
    -- The seeder and PUT /pengaturan both address rows by it, never by id.
    kunci        varchar(100)          NOT NULL UNIQUE,
    grup         varchar(50)           NOT NULL,
    label        varchar(150)          NOT NULL,
    -- Always stored as text; tipe_nilai says how to read it.
    nilai        text,
    tipe_nilai   tipe_nilai_pengaturan NOT NULL DEFAULT 'string',
    nilai_bawaan text,
    opsi         jsonb,
    satuan       varchar(20),
    keterangan   text,
    urutan       int                   NOT NULL DEFAULT 0,
    is_publik    boolean               NOT NULL DEFAULT false,
    is_terkunci  boolean               NOT NULL DEFAULT false,
    modified_at  timestamptz,
    modified_by  bigint
);

-- Drives the settings page: one tab per grup, rows in urutan order.
CREATE INDEX ix_pengaturan_grup ON mst_pengaturan (grup, urutan);

-- No soft-delete and no created_at: settings are seeded once and only ever
-- edited. Deleting one means deleting the feature that reads it.

CREATE TABLE log_aktivitas (
    id            bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- NULL for system/scheduled-job actions (nobody signed in did them).
    aktor_user_id bigint,
    modul         varchar(50)  NOT NULL,
    aksi          varchar(50)  NOT NULL,
    reff_type     varchar(50),
    reff_id       bigint,
    ringkasan     varchar(255),
    nilai_lama    jsonb,
    nilai_baru    jsonb,
    ip_address    inet,
    user_agent    text,
    created_at    timestamptz  NOT NULL DEFAULT now()
);

-- Append-only: no update, no delete, no soft-delete columns. Every read is
-- "latest first" within a filter, hence created_at DESC in each index.
CREATE INDEX ix_log_modul_waktu ON log_aktivitas (modul, created_at DESC);
CREATE INDEX ix_log_reff        ON log_aktivitas (reff_type, reff_id);
CREATE INDEX ix_log_aktor       ON log_aktivitas (aktor_user_id, created_at DESC);
