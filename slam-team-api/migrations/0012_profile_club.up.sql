-- 0012: profile_club — singleton club identity (banner, logos, address, tagline).
-- Table name is profile_club (NO mst_ prefix); GORM TableName() must return
-- exactly "profile_club" so GORM does not pluralise to "profile_clubs".

CREATE TABLE IF NOT EXISTS profile_club (
    id                   bigint       GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    nama                 varchar(150) NOT NULL,
    singkatan            varchar(50),
    banner_file_id       bigint,
    logo_simple_file_id  bigint,
    logo_besar_file_id   bigint,
    alamat               text,
    keterangan           text,

    -- Soft-delete columns (explicit, NOT gorm.DeletedAt).
    is_deleted           boolean      NOT NULL DEFAULT false,
    deleted_at           timestamptz,
    deleted_by           bigint,

    -- Audit columns.
    created_at           timestamptz  NOT NULL DEFAULT now(),
    created_by           bigint,
    modified_at          timestamptz,
    modified_by          bigint,

    -- FK → mst_file for three images.
    CONSTRAINT fk_profile_club_banner
        FOREIGN KEY (banner_file_id) REFERENCES mst_file(id) ON DELETE SET NULL,
    CONSTRAINT fk_profile_club_logo_simple
        FOREIGN KEY (logo_simple_file_id) REFERENCES mst_file(id) ON DELETE SET NULL,
    CONSTRAINT fk_profile_club_logo_besar
        FOREIGN KEY (logo_besar_file_id) REFERENCES mst_file(id) ON DELETE SET NULL
);

-- Index for the active-row guard (singleton check).
CREATE INDEX IF NOT EXISTS idx_profile_club_active
    ON profile_club (is_deleted) WHERE is_deleted = false;
