-- Migration 0005: mst_instansi — Master Data Instansi / Sekolah
-- Created: 2026-09-09

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
