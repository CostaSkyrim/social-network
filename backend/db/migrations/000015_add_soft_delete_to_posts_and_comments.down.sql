ALTER TABLE posts DROP COLUMN is_deleted;
ALTER TABLE posts DROP COLUMN deleted_at;
ALTER TABLE comments DROP COLUMN is_deleted;
ALTER TABLE comments DROP COLUMN deleted_at;