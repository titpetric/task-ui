-- user_auth_token: A token a client signs requests with, by name, kept as a hash
CREATE TABLE IF NOT EXISTS user_auth_token (
    id CHAR(26) PRIMARY KEY NOT NULL,
    user_id CHAR(26) NOT NULL,
    name TEXT NOT NULL,
    hash TEXT NOT NULL,
    used_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_auth_token_hash ON user_auth_token(hash);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_auth_token_user_name ON user_auth_token(user_id, name);
