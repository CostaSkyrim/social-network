CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_id INTEGER NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,

    -- Session management
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,

    -- Indexes for querying
    INDEX idx_sessions_session_id (session_id),
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_expires_at (expires_at)
);