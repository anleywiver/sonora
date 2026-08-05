-- Admin "Invite Member" (Sprint 14 sisipan) needs to create a user row
-- BEFORE they've ever logged in via Google — google_id isn't known yet
-- at invite time, only their email. NULL (not '') so the existing unique
-- index still allows any number of pending invites to coexist (Postgres
-- unique indexes treat NULLs as distinct from each other, unlike empty
-- strings which would collide on the second invite).
ALTER TABLE users ALTER COLUMN google_id DROP NOT NULL;
