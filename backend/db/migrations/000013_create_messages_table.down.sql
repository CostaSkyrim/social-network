DROP INDEX IF EXISTS idx_messages_is_read;
DROP INDEX IF EXISTS idx_messages_created_at;
DROP INDEX IF EXISTS idx_messages_sender;
DROP INDEX IF EXISTS idx_messages_group;
DROP INDEX IF EXISTS idx_messages_direct_message;

DROP TABLE IF EXISTS messages;
