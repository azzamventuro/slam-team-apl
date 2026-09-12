-- Migration 0014: mst_lokasi
-- Created: 2026-09-12
--
-- Master data lokasi: training/activity location with a geofence
-- (latitude/longitude + radius_meter) and an IANA timezone. jadwal.lokasi_id
-- will reference this table (restrict) once 16-jadwal lands.
-- users already exists (0005) so the audit FKs are added immediately.

CREATE TABLE mst_lokasi (
    id            bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode          varchar(50)   NOT NULL,
    nama          varchar(150)  NOT NULL,
    jenis_lokasi  varchar(50),
    alamat        text,
    latitude      decimal(10,7) NOT NULL,
    longitude     decimal(10,7) NOT NULL,
    radius_meter  int           NOT NULL DEFAULT 100,
    timezone      varchar(50)   NOT NULL DEFAULT 'Asia/Jakarta',
    foto_file_id  bigint,
    keterangan    text,
    is_aktif      boolean       NOT NULL DEFAULT true,
    is_deleted    boolean       NOT NULL DEFAULT false,
    deleted_at    timestamptz,
    deleted_by    bigint,
    created_at    timestamptz   NOT NULL DEFAULT now(),
    created_by    bigint,
    modified_at   timestamptz,
    modified_by   bigint,

    CONSTRAINT ck_mst_lokasi_latitude  CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT ck_mst_lokasi_longitude CHECK (longitude BETWEEN -180 AND 180),
    CONSTRAINT ck_mst_lokasi_radius    CHECK (radius_meter > 0)
);

-- Partial-unique index: kode must be unique among non-deleted rows.
CREATE UNIQUE INDEX ux_mst_lokasi_kode ON mst_lokasi (kode) WHERE is_deleted = false;

-- Filter by is_aktif.
CREATE INDEX ix_mst_lokasi_is_aktif ON mst_lokasi (is_aktif);

-- FK: foto file reference (nullable).
ALTER TABLE mst_lokasi ADD CONSTRAINT fk_lokasi_foto
    FOREIGN KEY (foto_file_id) REFERENCES mst_file (id) ON DELETE SET NULL;

-- FK: audit / soft-delete user references (nullable, immediate — users exists).
ALTER TABLE mst_lokasi ADD CONSTRAINT fk_lokasi_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_lokasi ADD CONSTRAINT fk_lokasi_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE mst_lokasi ADD CONSTRAINT fk_lokasi_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;
