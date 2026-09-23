DROP INDEX IF EXISTS idx_events_is_cancelled;

ALTER TABLE events DROP COLUMN is_cancelled;
