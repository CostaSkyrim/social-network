CREATE TABLE IF NOT EXISTS direct_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1_id INTEGER NOT NULL,
    user2_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMP,

    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE,

    UNIQUE(user1_id, user2_id)
);

CREATE INDEX IF NOT EXISTS idx_direct_messages_users ON direct_messages(user1_id, user2_id);
CREATE INDEX IF NOT EXISTS idx_direct_messages_last_message_at ON direct_messages(last_message_at);
