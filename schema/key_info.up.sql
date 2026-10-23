-- key_info: A key of the key folder, TASK_UI_SSH_DIR, by name: its files are <name>.key and <name>.pub there, and the row says what it is, an encryption or an SSH key, of the instance or a workspace
CREATE TABLE IF NOT EXISTS key_info (
    name TEXT PRIMARY KEY NOT NULL,
    purpose TEXT NOT NULL CHECK (purpose IN ('encrypt', 'ssh')),
    algo TEXT NOT NULL,
    bits INTEGER NOT NULL DEFAULT 0,
    fingerprint TEXT NOT NULL,
    scope TEXT NOT NULL DEFAULT 'instance',
    comment TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_key_info_fingerprint ON key_info(fingerprint);
CREATE INDEX IF NOT EXISTS idx_key_info_scope ON key_info(scope, purpose);
