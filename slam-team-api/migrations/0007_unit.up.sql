-- Migration 0007: unit — Master Data Unit (airsoft gun registry)
-- Created: 2026-09-10
-- Prerequisites: 0003_file_layer (mst_file), 0006_anggota (anggota).

CREATE TABLE unit (
    id                  bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    anggota_id          bigint        NOT NULL,
    kode                varchar(50),
    model               varchar(100),
    panjang             decimal(8,2),
    panjang_inbar       decimal(8,2),
    lebar               decimal(8,2),
    berat               decimal(8,2),
    berat_bb            decimal(8,3),
    fps                 decimal(8,2),
    deskripsi_warna     varchar(150),
    foto_sampul_file_id bigint,
    disetujui           boolean       NOT NULL DEFAULT false,
    disetujui_oleh      bigint,
    disetujui_pada      timestamptz,
    is_deleted          boolean       NOT NULL DEFAULT false,
    deleted_at          timestamptz,
    deleted_by          bigint,
    created_at          timestamptz   NOT NULL DEFAULT now(),
    created_by          bigint,
    modified_at         timestamptz,
    modified_by         bigint
);

-- Foreign keys
ALTER TABLE unit
    ADD CONSTRAINT fk_unit_anggota
        FOREIGN KEY (anggota_id) REFERENCES anggota(id),
    ADD CONSTRAINT fk_unit_foto_sampul
        FOREIGN KEY (foto_sampul_file_id) REFERENCES mst_file(id),
    ADD CONSTRAINT fk_unit_disetujui_oleh
        FOREIGN KEY (disetujui_oleh) REFERENCES users(id);

-- Performance indexes
CREATE INDEX IF NOT EXISTS idx_unit_anggota_id ON unit (anggota_id);
CREATE INDEX IF NOT EXISTS idx_unit_is_deleted ON unit (is_deleted) WHERE is_deleted = false;
