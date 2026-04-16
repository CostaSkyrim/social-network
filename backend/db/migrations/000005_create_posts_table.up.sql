CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    author_id INTEGER NOT NULL,
    group_id INTEGER,
    content TEXT,
    image_path VARCHAR(500),
    privacy_level VARCHAR(20) NOT NULL DEFAULT 'public',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,

    INDEX idx_posts_author_id (author_id),
    INDEX idx_posts_group_id (group_id),
    INDEX idx_posts_privacy_level (privacy_level),
    INDEX idx_posts_created_at (created_at)
);