-- key_user: A user, by name, whose runs get an SSH key of the key folder through the run's ssh agent
CREATE TABLE IF NOT EXISTS key_user (
    id CHAR(26) PRIMARY KEY NOT NULL,
    key_name TEXT NOT NULL,
    user_name TEXT NOT NULL,
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_key_user_key_user ON key_user(key_name, user_name);
