ALTER TABLE events ADD COLUMN reminder_sent BOOLEAN DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_events_reminder ON events(reminder_sent, event_datetime);
