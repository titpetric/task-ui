-- user_session: A browser's sign in, by the hash of its session cookie
CREATE TABLE IF NOT EXISTS user_session (
    id CHAR(26) PRIMARY KEY NOT NULL,
    user_id CHAR(26) NOT NULL,
    hash TEXT NOT NULL,
    expires_at DATETIME,
    used_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_session_hash ON user_session(hash);
CREATE INDEX IF NOT EXISTS idx_user_session_user_id ON user_session(user_id);
