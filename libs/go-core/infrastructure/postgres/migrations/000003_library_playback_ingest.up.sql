CREATE TABLE playlists (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    description  TEXT,
    cover_url    TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_playlists_user_id ON playlists(user_id);

-- position is fractional so reordering never requires rewriting every row.
CREATE TABLE playlist_songs (
    id           UUID PRIMARY KEY,
    playlist_id  UUID NOT NULL REFERENCES playlists(id) ON DELETE CASCADE,
    song_id      UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    position     DOUBLE PRECISION NOT NULL,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (playlist_id, song_id)
);

CREATE INDEX idx_playlist_songs_playlist_id ON playlist_songs(playlist_id, position);

-- Polymorphic favorite target (song/album/artist/playlist) per API design.
CREATE TABLE favorites (
    id                UUID PRIMARY KEY,
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    favoritable_type  TEXT NOT NULL CHECK (favoritable_type IN ('song', 'album', 'artist', 'playlist')),
    favoritable_id    UUID NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, favoritable_type, favoritable_id)
);

CREATE INDEX idx_favorites_user_id ON favorites(user_id);

CREATE TABLE play_history (
    id           UUID PRIMARY KEY,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    song_id      UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    progress_ms  INTEGER NOT NULL DEFAULT 0,
    played_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_play_history_user_id ON play_history(user_id, played_at DESC);

CREATE TABLE queue_items (
    id        UUID PRIMARY KEY,
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    song_id   UUID NOT NULL REFERENCES songs(id) ON DELETE CASCADE,
    position  DOUBLE PRECISION NOT NULL,
    added_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_queue_items_user_id ON queue_items(user_id, position);

-- One row per user: which device is the Active Device and what it's doing.
CREATE TABLE playback_states (
    user_id           UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    active_device_id  UUID REFERENCES devices(id) ON DELETE SET NULL,
    current_song_id   UUID REFERENCES songs(id) ON DELETE SET NULL,
    position_ms       INTEGER NOT NULL DEFAULT 0,
    is_playing        BOOLEAN NOT NULL DEFAULT false,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Auto-ingest sources are restricted to manual_upload, bandcamp, and
-- cloud_sync only — never scraping from licensed platforms. See CLAUDE.md.
CREATE TABLE ingest_jobs (
    id             UUID PRIMARY KEY,
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_type    TEXT NOT NULL CHECK (source_type IN ('manual_upload', 'bandcamp', 'cloud_sync')),
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    song_id        UUID REFERENCES songs(id) ON DELETE SET NULL,
    error_message  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ingest_jobs_user_id ON ingest_jobs(user_id, status);
