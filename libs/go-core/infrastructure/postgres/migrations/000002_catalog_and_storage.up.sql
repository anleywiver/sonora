-- Storage pool (Google Drive multi-account, Hetzner later behind the same
-- interface) must exist before songs, since every song points at the file
-- that backs it.

CREATE TABLE storage_accounts (
    id                     UUID PRIMARY KEY,
    provider               TEXT NOT NULL DEFAULT 'google_drive' CHECK (provider IN ('google_drive', 'hetzner')),
    label                  TEXT NOT NULL,
    account_email          TEXT,
    credentials_encrypted  TEXT NOT NULL,
    quota_bytes            BIGINT,
    used_bytes             BIGINT NOT NULL DEFAULT 0,
    is_active              BOOLEAN NOT NULL DEFAULT true,
    health_status          TEXT NOT NULL DEFAULT 'healthy' CHECK (health_status IN ('healthy', 'degraded', 'down')),
    last_health_check_at   TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE storage_files (
    id                  UUID PRIMARY KEY,
    storage_account_id  UUID NOT NULL REFERENCES storage_accounts(id) ON DELETE RESTRICT,
    provider_file_id    TEXT NOT NULL,
    checksum            TEXT NOT NULL,
    size_bytes          BIGINT NOT NULL,
    mime_type           TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (storage_account_id, provider_file_id)
);

CREATE INDEX idx_storage_files_checksum ON storage_files(checksum);

CREATE TABLE artists (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL,
    image_url   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE albums (
    id           UUID PRIMARY KEY,
    artist_id    UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    title        TEXT NOT NULL,
    cover_url    TEXT,
    released_at  DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_albums_artist_id ON albums(artist_id);

-- Checksum dedup happens before a song row is ever created (see ingest_jobs),
-- so checksum stays unique here as the final guard.
CREATE TABLE songs (
    id               UUID PRIMARY KEY,
    album_id         UUID REFERENCES albums(id) ON DELETE SET NULL,
    artist_id        UUID NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    storage_file_id  UUID NOT NULL REFERENCES storage_files(id) ON DELETE RESTRICT,
    title            TEXT NOT NULL,
    duration_ms      INTEGER NOT NULL,
    track_number     INTEGER,
    checksum         TEXT NOT NULL UNIQUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_songs_album_id ON songs(album_id);
CREATE INDEX idx_songs_artist_id ON songs(artist_id);
CREATE INDEX idx_songs_title ON songs(title);

CREATE TABLE genres (
    id    UUID PRIMARY KEY,
    name  TEXT NOT NULL UNIQUE
);

CREATE TABLE song_genres (
    song_id   UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    genre_id  UUID NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (song_id, genre_id)
);

CREATE TABLE lyrics_providers (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    base_url    TEXT NOT NULL,
    is_enabled  BOOLEAN NOT NULL DEFAULT true,
    priority    INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE lyrics (
    id               UUID PRIMARY KEY,
    song_id          UUID NOT NULL UNIQUE REFERENCES songs(id) ON DELETE CASCADE,
    provider_id      UUID REFERENCES lyrics_providers(id) ON DELETE SET NULL,
    synced_content   TEXT,
    plain_content    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
