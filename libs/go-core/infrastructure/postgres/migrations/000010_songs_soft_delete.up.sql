-- Admin Manage Songs (Sprint 14 sisipan, ADR 0010) — soft delete only;
-- scope of what respects this column is documented in the ADR.
ALTER TABLE songs ADD COLUMN deleted_at TIMESTAMPTZ;
