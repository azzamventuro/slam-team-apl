-- Migration 0013: sesi_login — refresh-token sessions (DBML line 400)
-- Created: 2026-09-12
--
-- One row per login; the opaque refresh token lives here (never a JWT), so a
-- session can be revoked server-side (logout, forced logout, rotation).
-- users already exists (0005) and is NOT touched by this migration.

CREATE TABLE sesi_login (
    id              bigint        GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id         bigint        NOT NULL,
    refresh_token   varchar(255)  NOT NULL,
    ip_address      inet,
    user_agent      text,
    info_perangkat  jsonb,
    berlaku_sampai  timestamptz   NOT NULL,
    dicabut_pada    timestamptz,
    created_at      timestamptz   NOT NULL DEFAULT now(),

    CONSTRAINT uq_sesi_login_refresh_token UNIQUE (refresh_token),
    CONSTRAINT fk_sesi_login_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
);

CREATE INDEX ix_sesi_login_user ON sesi_login (user_id);

COMMENT ON TABLE sesi_login IS
    'Memungkinkan logout paksa dan melihat perangkat aktif. Diperlukan karena PWA menyimpan sesi lebih lama.';
