CREATE TABLE mst_medsos (
    id              bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    anggota_id      bigint       NOT NULL,
    kode            varchar(50),
    tipe            int          NOT NULL DEFAULT 0,
    icon            varchar(50),
    jenis_medsos    varchar(50)  NOT NULL,
    konten_medsos   varchar(255) NOT NULL,
    is_deleted      boolean      NOT NULL DEFAULT false,
    deleted_at      timestamptz,
    deleted_by      bigint,
    created_at      timestamptz  NOT NULL DEFAULT now(),
    created_by      bigint,
    modified_at     timestamptz,
    modified_by     bigint,

    CONSTRAINT fk_mst_medsos_anggota
        FOREIGN KEY (anggota_id) REFERENCES anggota(id)
);

CREATE INDEX idx_mst_medsos_anggota_id
    ON mst_medsos(anggota_id)
    WHERE is_deleted = false;
