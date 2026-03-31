CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    nickname VARCHAR(100) UNIQUE,
    date_of_birth DATE NOT NULL,
    about_me TEXT,
    avatar_path VARCHAR(500),

    -- Privacy setting
    is_public BOOLEAN DEFAULT TRUE,

    -- Account status
    is_active BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes for querying
    INDEX idx_users_email (email),
    INDEX idx_users_uuid (uuid),
    INDEX idx_users_nickname (nickname),
    INDEX idx_users_created_at (created_at)
);