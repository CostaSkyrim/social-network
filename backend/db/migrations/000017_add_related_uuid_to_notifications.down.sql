DROP INDEX IF EXISTS idx_notifications_related_uuid;

ALTER TABLE notifications DROP COLUMN related_uuid;
