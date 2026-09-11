-- 0008_prestasi.up.sql
-- Creates the prestasi table for competition achievements.

CREATE TABLE prestasi (
    id                  bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    anggota_id          bigint       NOT NULL,
    kode                varchar(50),
    peringkat           varchar(50),
    tingkat             varchar(50),
    judul_kompetisi     varchar(200),
    flyer_file_id       bigint,
    tanggal_kompetisi   date,
    alamat_kompetisi    text,
    foto_sampul_file_id bigint,
    keterangan          text,
    is_deleted          boolean      NOT NULL DEFAULT false,
    deleted_at          timestamptz,
    deleted_by          bigint,
    created_at          timestamptz  NOT NULL DEFAULT now(),
    created_by          bigint,
    modified_at         timestamptz,
    modified_by         bigint
);

-- FK constraints.
ALTER TABLE prestasi ADD CONSTRAINT fk_prestasi_anggota
    FOREIGN KEY (anggota_id) REFERENCES anggota(id);
ALTER TABLE prestasi ADD CONSTRAINT fk_prestasi_flyer_file
    FOREIGN KEY (flyer_file_id) REFERENCES mst_file(id);
ALTER TABLE prestasi ADD CONSTRAINT fk_prestasi_foto_sampul_file
    FOREIGN KEY (foto_sampul_file_id) REFERENCES mst_file(id);

-- Query indexes.
CREATE INDEX idx_prestasi_anggota_id ON prestasi (anggota_id);
CREATE INDEX idx_prestasi_tanggal ON prestasi (tanggal_kompetisi);
CREATE INDEX idx_prestasi_not_deleted ON prestasi (id) WHERE is_deleted = false;
