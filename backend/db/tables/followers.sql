CREATE TABLE IF NOT EXISTS followers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    follower_id INTEGER NOT NULL,
    following_id INTEGER NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- can be pending, accepted, declined

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Prevent duplicate follow relationships
    UNIQUE(follower_id, following_id),

    -- Indexes for querying
    INDEX idx_followers_follower (follower_id),
    INDEX idx_followers_following (following_id),
    INDEX idx_followers_status (status)
);