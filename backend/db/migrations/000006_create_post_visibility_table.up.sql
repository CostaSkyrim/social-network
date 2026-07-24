CREATE TABLE IF NOT EXISTS post_visibility (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    UNIQUE(post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_post_visibility_post ON post_visibility(post_id);
CREATE INDEX IF NOT EXISTS idx_post_visibility_user ON post_visibility(user_id);
