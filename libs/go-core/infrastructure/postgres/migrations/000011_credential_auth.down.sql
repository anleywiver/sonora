DROP TABLE IF EXISTS app_settings;
ALTER TABLE users DROP COLUMN password_hash;
ALTER TABLE users DROP COLUMN username;
