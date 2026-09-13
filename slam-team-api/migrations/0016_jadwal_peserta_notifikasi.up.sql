-- Migration 0016: jadwal_peserta + notifikasi
-- Created: 2026-09-13
--
-- jadwal_peserta assigns an anggota (NOT a user — members without an account
-- can still be listed as participants) to a schedule, optionally to one
-- session (sesi_id NULL = every session of the schedule). wajib_absen is
-- forced false by the service for account-less anggota so the session-closing
-- job never marks them alfa: the rule lives in the data, not in a query.
--
-- notifikasi is the in-app inbox, fanned out one row per recipient: assigning
-- 50 members writes 50 rows, because read state belongs to each person and
-- the (user_id, is_dibaca, created_at) index keeps the badge query cheap.
-- Only account holders receive rows (user_id NOT NULL → users, CASCADE).
--
-- Enum types status_tugas / prioritas_notifikasi were created in 0001_enums.
-- jadwal + jadwal_sesi (0015), anggota (0006) and users (0005) exist, so every
-- FK is added here immediately.

CREATE TABLE jadwal_peserta (
    id               bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    jadwal_id        bigint        NOT NULL,
    sesi_id          bigint,                                        -- NULL = berlaku semua sesi jadwal ini
    anggota_id       bigint        NOT NULL,                        -- penugasan memakai anggota, BUKAN user
    peran_peserta    varchar(50),                                   -- peserta / pelatih / panitia / pengawas
    wajib_absen      boolean       NOT NULL DEFAULT true,           -- otomatis false bila anggota tidak punya akun user
    status_tugas     status_tugas  NOT NULL DEFAULT 'ditugaskan',
    ditugaskan_oleh  bigint        NOT NULL,
    ditugaskan_pada  timestamptz   NOT NULL DEFAULT now(),
    direspon_pada    timestamptz,
    keterangan       text,
    is_deleted       boolean       NOT NULL DEFAULT false,
    deleted_at       timestamptz,
    deleted_by       bigint
);

-- Indexes per .dbml.
CREATE INDEX ix_jadwal_peserta_anggota_jadwal ON jadwal_peserta (anggota_id, jadwal_id);
CREATE INDEX ix_jadwal_peserta_jadwal_sesi    ON jadwal_peserta (jadwal_id, sesi_id);

-- FK: parent schedule / session / member (restrict — an assignment is a fact
-- about a schedule and a member; neither may vanish underneath it).
ALTER TABLE jadwal_peserta ADD CONSTRAINT fk_jadwal_peserta_jadwal
    FOREIGN KEY (jadwal_id) REFERENCES jadwal (id) ON DELETE RESTRICT;
ALTER TABLE jadwal_peserta ADD CONSTRAINT fk_jadwal_peserta_sesi
    FOREIGN KEY (sesi_id) REFERENCES jadwal_sesi (id);
ALTER TABLE jadwal_peserta ADD CONSTRAINT fk_jadwal_peserta_anggota
    FOREIGN KEY (anggota_id) REFERENCES anggota (id) ON DELETE RESTRICT;

-- FK: assigner + soft-delete actor (users exists).
ALTER TABLE jadwal_peserta ADD CONSTRAINT fk_jadwal_peserta_ditugaskan_oleh
    FOREIGN KEY (ditugaskan_oleh) REFERENCES users (id);
ALTER TABLE jadwal_peserta ADD CONSTRAINT fk_jadwal_peserta_deleted_by
    FOREIGN KEY (deleted_by) REFERENCES users (id) ON DELETE SET NULL;

-- ─────────────────────────────────────────────────────────────────────────────

CREATE TABLE notifikasi (
    id               bigint                GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid             uuid                  NOT NULL DEFAULT gen_random_uuid(),
    user_id          bigint                NOT NULL,                -- penerima; hanya pemilik akun
    tipe             varchar(50)           NOT NULL,                -- jadwal_ditugaskan / jadwal_pengingat / sesi_dibatalkan / izin_* / absensi_dioverride / kta_* / sistem
    judul            varchar(200)          NOT NULL,
    isi              text,
    ikon             varchar(50),
    warna            varchar(20),                                   -- info / sukses / peringatan / bahaya
    route            varchar(255),                                  -- tujuan saat diklik, mis. /jadwal/12/sesi/340
    reff_type        varchar(50),                                   -- entitas sumber — polymorphic
    reff_id          bigint,
    prioritas        prioritas_notifikasi  NOT NULL DEFAULT 'normal',
    is_dibaca        boolean               NOT NULL DEFAULT false,
    dibaca_pada      timestamptz,
    is_diarsipkan    boolean               NOT NULL DEFAULT false,
    kedaluwarsa_pada timestamptz,                                   -- dibersihkan job terjadwal setelah sesi lewat
    created_at       timestamptz           NOT NULL DEFAULT now(),
    created_by       bigint,                                        -- pemicu; NULL untuk notifikasi sistem

    CONSTRAINT uq_notifikasi_uuid UNIQUE (uuid)
);

-- Indexes per .dbml. The first one serves both the unread badge and the list.
CREATE INDEX ix_notifikasi_user_dibaca_created ON notifikasi (user_id, is_dibaca, created_at);
CREATE INDEX ix_notifikasi_reff                ON notifikasi (reff_type, reff_id);
CREATE INDEX ix_notifikasi_kedaluwarsa         ON notifikasi (kedaluwarsa_pada);

-- FK: recipient (CASCADE — an inbox has no meaning without its owner) and the
-- triggering user (SET NULL — the row is still meaningful to the recipient).
ALTER TABLE notifikasi ADD CONSTRAINT fk_notifikasi_user
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;
ALTER TABLE notifikasi ADD CONSTRAINT fk_notifikasi_created_by
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL;
