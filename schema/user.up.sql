-- user: A login, with the role that decides what it may do
CREATE TABLE IF NOT EXISTS user (
    id CHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    password TEXT NOT NULL DEFAULT '',
    role TEXT NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'runner', 'viewer')),
    folders TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_name ON user(name);
