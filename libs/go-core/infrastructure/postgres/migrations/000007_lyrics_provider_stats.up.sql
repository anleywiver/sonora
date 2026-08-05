-- Sprint 14: admin Lyrics Source page needs Health + Match Rate columns
-- (docs/screens-spec.md #19) that never existed — lookups were never
-- counted before this (a miss left no trace at all).
ALTER TABLE lyrics_providers ADD COLUMN health_status TEXT NOT NULL DEFAULT 'online' CHECK (health_status IN ('online', 'rate_limited'));
ALTER TABLE lyrics_providers ADD COLUMN total_lookups BIGINT NOT NULL DEFAULT 0;
ALTER TABLE lyrics_providers ADD COLUMN successful_matches BIGINT NOT NULL DEFAULT 0;
