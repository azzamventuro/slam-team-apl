-- Migration 0017: absensi_izin
-- Created: 2026-09-13
--
-- absensi_izin is a member's request to be excused from one session:
-- izin / sakit / dinas (absent with reason) or pulang_cepat (leaving early,
-- with the requested clock time). A request is tied to a jadwal_sesi and an
-- anggota — both RESTRICT, because an approved izin is the evidence that the
-- member was not alfa, and that evidence must not vanish with its parents.
-- jadwal_id is denormalised from the session so the list can filter by
-- schedule without a join.
--
-- The reverse link (absensi.izin_id → absensi_izin) is added by the absensi
-- migration, not here: the absensi table does not exist yet.
--
-- Enum types jenis_izin / status_izin were created in 0001_enums.
-- jadwal + jadwal_sesi (0015), anggota (0006), mst_file (0003) and users
-- (0005) exist, so every FK is added here immediately.

CREATE TABLE absensi_izin (
    id                   bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    sesi_id              bigint       NOT NULL,
    jadwal_id            bigint,                                    -- disalin dari jadwal_sesi.jadwal_id
    anggota_id           bigint       NOT NULL,
    jenis                jenis_izin   NOT NULL,                     -- izin / sakit / dinas / pulang_cepat
    alasan               text         NOT NULL,
    waktu_pulang_diminta time,                                      -- khusus jenis = pulang_cepat
    lampiran_file_id     bigint,                                    -- mis. surat dokter. PRIVAT.
    status               status_izin  NOT NULL DEFAULT 'menunggu',  -- menunggu / disetujui / ditolak
    diproses_oleh        bigint,
    diproses_pada        timestamptz,
    catatan_peninjau     text,                                      -- wajib saat tolak
    is_deleted           boolean      NOT NULL DEFAULT false,
    deleted_at           timestamptz,
    deleted_by           bigint,
    created_at           timestamptz  NOT NULL DEFAULT now(),
    created_by           bigint,
    modified_at          timestamptz,
    modified_by          bigint
);

-- Indexes per .dbml: the first serves "does this member already have a
-- request for this session", the second the approval queue (status + age).
CREATE INDEX ix_absensi_izin_sesi_anggota    ON absensi_izin (sesi_id, anggota_id);
CREATE INDEX ix_absensi_izin_status_created  ON absensi_izin (status, created_at);

-- FK: session + member (restrict — see header), schedule, attachment, reviewer.
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_sesi
    FOREIGN KEY (sesi_id) REFERENCES jadwal_sesi (id) ON DELETE RESTRICT;
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_jadwal
    FOREIGN KEY (jadwal_id) REFERENCES jadwal (id);
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_anggota
    FOREIGN KEY (anggota_id) REFERENCES anggota (id) ON DELETE RESTRICT;
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_lampiran
    FOREIGN KEY (lampiran_file_id) REFERENCES mst_file (id);
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_diproses_oleh
    FOREIGN KEY (diproses_oleh) REFERENCES users (id);

-- FK: audit actors (users exists).
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_modified_by
    FOREIGN KEY (modified_by) REFERENCES users (id) ON DELETE SET NULL;
ALTER TABLE absensi_izin ADD CONSTRAINT fk_absensi_izin_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;
