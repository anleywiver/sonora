-- Sprint 11: waveform peaks + MusicBrainz enrichment (best-effort, see
-- docs/decisions/0005-sprint11-waveform-musicbrainz-analytics.md). All
-- nullable — ingest still completes fine without either.

ALTER TABLE songs ADD COLUMN waveform_peaks SMALLINT[];
ALTER TABLE songs ADD COLUMN musicbrainz_id TEXT;
ALTER TABLE artists ADD COLUMN musicbrainz_id TEXT;
ALTER TABLE albums ADD COLUMN musicbrainz_id TEXT;
