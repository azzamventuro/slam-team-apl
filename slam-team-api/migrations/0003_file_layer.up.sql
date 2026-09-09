-- 0003_file_layer — "Kelompok 4 — File" from slamteam_db.dbml: the centralised
-- upload layer, built once in Fase 0 before any module touches a file.
--
-- Every other module stores an image or document as a `*_file_id bigint`
-- pointing at mst_file.id — never as a path column of its own. The resize /
-- auto-orient / variant logic therefore exists in exactly one place.
--
-- ORDER: after 0002_pengaturan_log, because the variant pixel sizes are read
-- from mst_pengaturan grup "file" (file.varian_original_maks_px / _medium_px /
-- _low_px) rather than hardcoded.
--
-- ENUMS: varian_file and status_proses_file are created by 0001_enums.
--
-- FOREIGN KEYS: .dbml declares created_by / modified_by / deleted_by > users.id.
-- The users table belongs to a LATER migration (auth/identity), so the columns
-- are created plain here and the constraints are added by that migration — a FK
-- cannot precede its target. Same decision as 0002_pengaturan_log.

CREATE TABLE mst_file (
    id             bigint             GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- Public identity used in URLs. `id` never leaves the server.
    uuid           uuid               NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    -- Display only. The name on disk is always "{varian}.{ext}".
    nama_asli      varchar(255),
    -- Used in the readable URL: {modul}/{reff_id}/{nama_slug}-{varian}.{ext}
    nama_slug      varchar(255),
    ekstensi       varchar(10),
    -- Detected from the first bytes of the content, never from the extension.
    mime_type      varchar(100),
    ukuran_byte    bigint,
    -- Dedup + integrity proof for absensi evidence.
    hash_sha256    char(64),
    -- NULL for non-images (a PDF has no pixel size).
    lebar_px       int,
    tinggi_px      int,
    kategori       varchar(50)        NOT NULL,
    -- Polymorphic owner. reff_id may be filled after the owning row exists.
    reff_type      varchar(50),
    reff_id        bigint,
    storage_driver varchar(20)        NOT NULL DEFAULT 'local',
    -- Folder that maps URL to disk: {kategori}/{tahun}/{bulan}/{uuid}
    path_dasar     varchar(500),
    -- false = MUST be served by the authenticated Go handler, never by nginx.
    is_publik      boolean            NOT NULL DEFAULT false,
    -- menunggu → selesai | gagal. Variants are generated asynchronously, and a
    -- failure has to stay visible rather than silently look like success.
    status_proses  status_proses_file NOT NULL DEFAULT 'menunggu',
    -- Original EXIF kept as evidence; re-encoding strips it from the variants.
    metadata_exif  jsonb,
    is_deleted     boolean            NOT NULL DEFAULT false,
    deleted_at     timestamptz,
    deleted_by     bigint,
    created_at     timestamptz        NOT NULL DEFAULT now(),
    created_by     bigint,
    modified_at    timestamptz,
    modified_by    bigint
);

-- "every file of this anggota / this absensi" — the owning module's lookup.
CREATE INDEX ix_file_reff     ON mst_file (reff_type, reff_id);
-- Dedup probe on upload.
CREATE INDEX ix_file_hash     ON mst_file (hash_sha256);
CREATE INDEX ix_file_kategori ON mst_file (kategori);

CREATE TABLE mst_file_varian (
    id          bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    file_id     bigint       NOT NULL,
    varian      varian_file  NOT NULL,
    -- Relative to the storage root, so moving the root does not rewrite rows.
    path        varchar(500) NOT NULL,
    -- Ready-to-use URL, only meaningful when mst_file.is_publik.
    url_publik  text,
    lebar_px    int,
    tinggi_px   int,
    ukuran_byte bigint,
    -- What was actually written (image/jpeg), not what was requested.
    mime_type   varchar(100),
    kualitas    int,
    created_at  timestamptz  NOT NULL DEFAULT now(),

    CONSTRAINT fk_filevarian_file FOREIGN KEY (file_id)
        REFERENCES mst_file (id) ON DELETE CASCADE,
    -- One row per (file, varian): 3 for an image, 1 for a document.
    CONSTRAINT uq_file_varian UNIQUE (file_id, varian)
);
