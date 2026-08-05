-- Filter rules for auto-ingest sources ONLY (bandcamp/cloud_sync) — never
-- applies to manual_upload (CLAUDE.md: the user can always upload
-- anything themselves). See docs/decisions/0008-ingest-filter-rules.md.
CREATE TABLE ingest_filter_rules (
    id UUID PRIMARY KEY,
    source_type TEXT NOT NULL CHECK (source_type IN ('bandcamp', 'cloud_sync')),
    rule_type TEXT NOT NULL CHECK (rule_type IN ('genre_allow', 'year_min', 'year_max')),
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ingest_filter_rules_source_type_idx ON ingest_filter_rules (source_type);

-- 'processing' is kept as-is (existing jobs and code already use that
-- exact value since Sprint 3) — only adding the two new terminal states.
ALTER TABLE ingest_jobs DROP CONSTRAINT ingest_jobs_status_check;
ALTER TABLE ingest_jobs ADD CONSTRAINT ingest_jobs_status_check
    CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'needs_manual_upload', 'skipped_by_filter'));
