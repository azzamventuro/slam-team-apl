-- 0009_inorga.up.sql
-- Creates mst_inorga table (kepengurusan / organisational periods).

CREATE TABLE IF NOT EXISTS mst_inorga (
    id              BIGSERIAL       PRIMARY KEY,
    kode            VARCHAR(50),
    nama            VARCHAR(150)    NOT NULL,
    logo_file_id    BIGINT          REFERENCES mst_file(id) ON DELETE SET NULL,
    banner_file_id  BIGINT          REFERENCES mst_file(id) ON DELETE SET NULL,
    tanggal_mulai   DATE,
    tanggal_selesai DATE,
    file_sk_file_id BIGINT          REFERENCES mst_file(id) ON DELETE SET NULL,
    konten          TEXT,
    is_deleted      BOOLEAN         NOT NULL DEFAULT false,
    deleted_at      TIMESTAMPTZ,
    deleted_by      BIGINT,
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    created_by      BIGINT,
    modified_at     TIMESTAMPTZ,
    modified_by     BIGINT
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mst_inorga_is_deleted ON mst_inorga(is_deleted) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_mst_inorga_tanggal_mulai ON mst_inorga(tanggal_mulai);
