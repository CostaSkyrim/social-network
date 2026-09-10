DROP INDEX IF EXISTS idx_events_reminder;

ALTER TABLE events DROP COLUMN reminder_sent;
