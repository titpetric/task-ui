-- user_auth_password: A password a user signs in with, kept apart from the user row
CREATE TABLE IF NOT EXISTS user_auth_password (
    id CHAR(26) PRIMARY KEY NOT NULL,
    user_id CHAR(26) NOT NULL,
    hash TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'password' CHECK (kind IN ('password', 'backup')),
    used_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_user_auth_password_user_id ON user_auth_password(user_id, kind);

-- The hash moves here; the password file fills this table on startup.
ALTER TABLE user DROP COLUMN password;
