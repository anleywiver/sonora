-- temp_path holds the on-disk path of the uploaded file while a job is
-- pending/processing, so it can be retried after a failure without asking
-- the user to re-upload. Cleared (and the file deleted) once the job
-- reaches 'completed' since the content is durably stored in Drive by then.
ALTER TABLE ingest_jobs ADD COLUMN temp_path TEXT;
