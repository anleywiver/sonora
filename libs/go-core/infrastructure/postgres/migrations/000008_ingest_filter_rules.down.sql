ALTER TABLE ingest_jobs DROP CONSTRAINT ingest_jobs_status_check;
ALTER TABLE ingest_jobs ADD CONSTRAINT ingest_jobs_status_check
    CHECK (status IN ('pending', 'processing', 'completed', 'failed'));

DROP TABLE IF EXISTS ingest_filter_rules;
