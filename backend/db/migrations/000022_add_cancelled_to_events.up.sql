ALTER TABLE events ADD COLUMN is_cancelled BOOLEAN DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_events_is_cancelled ON events(is_cancelled);
