CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    sender_id INTEGER NOT NULL,
    direct_message_id INTEGER,
    group_id INTEGER,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (direct_message_id) REFERENCES direct_messages(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,

    CHECK ((direct_message_id IS NOT NULL AND group_id IS NULL) OR
           (direct_message_id IS NULL AND group_id IS NOT NULL)),

    INDEX idx_messages_direct_message (direct_message_id),
    INDEX idx_messages_group (group_id),
    INDEX idx_messages_sender (sender_id),
    INDEX idx_messages_created_at (created_at),
    INDEX idx_messages_is_read (is_read)
);