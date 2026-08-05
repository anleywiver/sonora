-- Lossy on purpose: a NOT NULL rollback can't coexist with pending
-- (never-logged-in) invites, so they're removed first.
DELETE FROM users WHERE google_id IS NULL;
ALTER TABLE users ALTER COLUMN google_id SET NOT NULL;
