-- Sprint 14 sisipan (ADR 0012) — credential-based (username/password)
-- login becomes the default auth path; Google OAuth stays fully
-- functional, gated by a runtime toggle in app_settings (not removed,
-- not hardcoded off). Both nullable — a Google-only user never has a
-- password, a not-yet-created-by-admin path never applies to existing
-- Google users.
ALTER TABLE users ADD COLUMN username TEXT UNIQUE;
ALTER TABLE users ADD COLUMN password_hash TEXT;

-- Generic key-value settings store — also backs the Sprint 14 sisipan
-- Admin Settings page (app_name, default_language, maintenance_mode)
-- from the earlier sisipan request, not just the OAuth toggle.
CREATE TABLE app_settings (
    key         TEXT PRIMARY KEY,
    value       TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app_settings (key, value) VALUES
    ('google_oauth_enabled', 'false'),
    ('maintenance_mode', 'false'),
    ('app_name', 'Sonora'),
    ('default_language', 'id');
