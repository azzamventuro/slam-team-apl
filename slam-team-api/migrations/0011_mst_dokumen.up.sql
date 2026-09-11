-- 0011_mst_dokumen — Polymorphic document attachments.
-- ORDER: after 0003_file_layer (mst_file FK).

CREATE TABLE mst_dokumen (
    id           bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode         varchar(50),
    file_id      bigint,
    tipe         varchar(50),
    format       varchar(20),
    reff_id      bigint       NOT NULL DEFAULT 0,
    reff_type    varchar(50)  NOT NULL,
    jenis        int,
    keterangan   varchar(255),
    is_deleted   boolean      NOT NULL DEFAULT false,
    deleted_at   timestamptz,
    deleted_by   bigint,
    created_at   timestamptz  NOT NULL DEFAULT now(),
    created_by   bigint,
    modified_at  timestamptz,
    modified_by  bigint,

    CONSTRAINT fk_mst_dokumen_file
        FOREIGN KEY (file_id) REFERENCES mst_file(id) ON DELETE SET NULL
);

-- Composite index for the most common query: "all documents for one entity".
CREATE INDEX idx_mst_dokumen_reff
    ON mst_dokumen (reff_type, reff_id)
    WHERE is_deleted = false;
