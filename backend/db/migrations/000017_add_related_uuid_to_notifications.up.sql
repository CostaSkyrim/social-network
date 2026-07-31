ALTER TABLE notifications ADD COLUMN related_uuid VARCHAR(36);

CREATE INDEX idx_notifications_related_uuid ON notifications(related_uuid);
