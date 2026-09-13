-- Migration 0015: jadwal + jadwal_sesi
-- Created: 2026-09-13
--
-- jadwal is the schedule definition: a recurrence rule (pola_ulang /
-- hari_ulang / interval_ulang / tanggal_akhir_ulang) plus the attendance rules
-- (mode_absen, buka/tutup window, geofence snapshot). jadwal_sesi is one
-- concrete date materialised from that rule by POST /jadwal/{id}/generate-sesi
-- and is the row absensi binds to — a single-date schedule still gets exactly
-- one jadwal_sesi row so the attendance flow is uniform.
--
-- Enum types pola_ulang / mode_absen / status_sesi were created in 0001_enums.
-- mst_lokasi (0014), mst_inorga (0009) and users (0005) exist, so those FKs are
-- added here. kegiatan does NOT exist yet (Fase 8) — the kegiatan_id column is
-- created now and its FK is deferred; see the seam marker at the bottom.

CREATE TABLE jadwal (
    id                   bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    kode                 varchar(50)   NOT NULL,
    nama                 varchar(150)  NOT NULL,
    jenis_jadwal         varchar(50),
    deskripsi            text,
    lokasi_id            bigint,
    latitude             decimal(10,7),                                 -- snapshot dari mst_lokasi saat jadwal dibuat
    longitude            decimal(10,7),
    radius_meter         int           NOT NULL DEFAULT 100,            -- menimpa radius lokasi
    timezone             varchar(50)   NOT NULL DEFAULT 'Asia/Jakarta', -- nama IANA
    tanggal_mulai        date          NOT NULL,
    tanggal_selesai      date,
    jam_mulai            time          NOT NULL,
    jam_selesai          time          NOT NULL,
    is_berulang          boolean       NOT NULL DEFAULT false,
    pola_ulang           pola_ulang    NOT NULL DEFAULT 'tidak_berulang',
    hari_ulang           int[],                                         -- pola mingguan: {0..6}, 0 = Minggu
    interval_ulang       int           NOT NULL DEFAULT 1,
    tanggal_akhir_ulang  date,
    mode_absen           mode_absen    NOT NULL DEFAULT 'masuk_saja',
    wajib_absen          boolean       NOT NULL DEFAULT true,
    butuh_selfie         boolean       NOT NULL DEFAULT true,
    butuh_lokasi         boolean       NOT NULL DEFAULT true,
    izinkan_luar_radius  boolean       NOT NULL DEFAULT true,
    toleransi_telat_mnt  int           NOT NULL DEFAULT 15,
    buka_absen_mnt       int           NOT NULL DEFAULT 30,             -- menit sebelum mulai
    tutup_absen_mnt      int           NOT NULL DEFAULT 60,             -- menit setelah mulai
    kuota                int,
    kegiatan_id          bigint,                                        -- FK deferred (see bottom)
    inorga_id            bigint,
    status               varchar(20)   NOT NULL DEFAULT 'draft',        -- bukan enum: draft / terbit / dibatalkan / selesai
    is_deleted           boolean       NOT NULL DEFAULT false,
    deleted_at           timestamptz,
    deleted_by           bigint,
    created_at           timestamptz   NOT NULL DEFAULT now(),
    created_by           bigint,
    modified_at          timestamptz,
    modified_by          bigint,

    CONSTRAINT ck_jadwal_jam        CHECK (jam_selesai > jam_mulai),
    CONSTRAINT ck_jadwal_status     CHECK (status IN ('draft', 'terbit', 'dibatalkan', 'selesai')),
    CONSTRAINT ck_jadwal_hari_ulang CHECK (hari_ulang IS NULL OR hari_ulang <@ ARRAY[0, 1, 2, 3, 4, 5, 6]),
    CONSTRAINT ck_jadwal_interval   CHECK (interval_ulang > 0),
    CONSTRAINT ck_jadwal_radius     CHECK (radius_meter > 0),
    CONSTRAINT ck_jadwal_latitude   CHECK (latitude  IS NULL OR latitude  BETWEEN -90  AND 90),
    CONSTRAINT ck_jadwal_longitude  CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180),
    CONSTRAINT ck_jadwal_absen_mnt  CHECK (toleransi_telat_mnt >= 0 AND buka_absen_mnt >= 0 AND tutup_absen_mnt >= 0)
);

-- Partial-unique index: kode must be unique among non-deleted rows.
CREATE UNIQUE INDEX ux_jadwal_kode ON jadwal (kode) WHERE is_deleted = false;

-- Indexes per .dbml.
CREATE INDEX ix_jadwal_tanggal_mulai_status ON jadwal (tanggal_mulai, status);
CREATE INDEX ix_jadwal_lokasi_id            ON jadwal (lokasi_id);

-- FK: lokasi (restrict — a location in use by a schedule cannot be removed).
ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_lokasi
    FOREIGN KEY (lokasi_id) REFERENCES mst_lokasi (id) ON DELETE RESTRICT;

-- FK: inorga (nullable).
ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_inorga
    FOREIGN KEY (inorga_id) REFERENCES mst_inorga (id) ON DELETE SET NULL;

-- FK: audit / soft-delete user references (nullable, immediate — users exists).
ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;

-- TODO(22-kegiatan): the kegiatan table is created in Fase 8. When it lands,
-- add in that migration:
--   ALTER TABLE jadwal ADD CONSTRAINT fk_jadwal_kegiatan
--       FOREIGN KEY (kegiatan_id) REFERENCES kegiatan (id) ON DELETE SET NULL;

-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE jadwal_sesi (
    id               bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    jadwal_id        bigint       NOT NULL,
    tanggal_lokal    date         NOT NULL,                       -- tanggal di zona jadwal
    mulai_utc        timestamptz  NOT NULL,                       -- jam lokal → UTC memakai jadwal.timezone
    selesai_utc      timestamptz  NOT NULL,
    timezone         varchar(50)  NOT NULL,                       -- disalin dari jadwal
    absen_buka_utc   timestamptz  NOT NULL,                       -- jendela absen
    absen_tutup_utc  timestamptz  NOT NULL,
    status           status_sesi  NOT NULL DEFAULT 'terjadwal',
    alasan_batal     text,
    jml_ditugaskan   int          NOT NULL DEFAULT 0,
    jml_hadir        int          NOT NULL DEFAULT 0,
    catatan          text,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    created_by       bigint,
    modified_at      timestamptz,
    modified_by      bigint,

    -- Idempotency anchor for generate-sesi (INSERT … ON CONFLICT DO NOTHING).
    CONSTRAINT uq_jadwal_sesi_jadwal_tanggal UNIQUE (jadwal_id, tanggal_lokal),
    CONSTRAINT ck_jadwal_sesi_waktu   CHECK (selesai_utc > mulai_utc),
    CONSTRAINT ck_jadwal_sesi_jendela CHECK (absen_tutup_utc > absen_buka_utc)
);

-- Indexes per .dbml.
CREATE INDEX ix_jadwal_sesi_tanggal_status ON jadwal_sesi (tanggal_lokal, status);
CREATE INDEX ix_jadwal_sesi_jendela        ON jadwal_sesi (absen_buka_utc, absen_tutup_utc);

-- FK: parent schedule (restrict — sessions are the absensi anchor).
ALTER TABLE jadwal_sesi ADD CONSTRAINT fk_jadwal_sesi_jadwal
    FOREIGN KEY (jadwal_id) REFERENCES jadwal (id) ON DELETE RESTRICT;

-- FK: audit user references (nullable, immediate — users exists).
ALTER TABLE jadwal_sesi ADD CONSTRAINT fk_jadwal_sesi_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE jadwal_sesi ADD CONSTRAINT fk_jadwal_sesi_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
