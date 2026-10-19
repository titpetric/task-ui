-- ssh_key_user: A user whose runs get an instance SSH key through the run's ssh agent, by name
CREATE TABLE IF NOT EXISTS ssh_key_user (
    id CHAR(26) PRIMARY KEY NOT NULL,
    key_name TEXT NOT NULL,
    user_name TEXT NOT NULL,
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_user_key_user ON ssh_key_user(key_name, user_name);
CREATE INDEX IF NOT EXISTS idx_ssh_key_user_user ON ssh_key_user(user_name);
