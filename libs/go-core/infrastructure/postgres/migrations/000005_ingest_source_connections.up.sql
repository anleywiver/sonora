-- Sprint 10: Bandcamp/cloud sync ingest sources, managed by Owner (see
-- docs/decisions/0004-sprint10-scheduled-jobs-and-ingest-sources.md).
-- Shape mirrors storage_accounts on purpose (credentials_encrypted via the
-- same crypto.Box, last_synced_at instead of last_health_check_at).

CREATE TABLE ingest_source_connections (
    id                     UUID PRIMARY KEY,
    provider               TEXT NOT NULL CHECK (provider IN ('bandcamp', 'cloud_sync')),
    label                  TEXT NOT NULL,
    account_email          TEXT,
    credentials_encrypted  TEXT NOT NULL,
    is_active              BOOLEAN NOT NULL DEFAULT true,
    last_synced_at         TIMESTAMPTZ,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);
