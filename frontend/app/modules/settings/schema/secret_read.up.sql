-- secret_read: A secret a run got in its environment, one row per secret per run, the last 1000 of each secret kept
CREATE TABLE IF NOT EXISTS secret_read (
    id CHAR(26) PRIMARY KEY NOT NULL,
    secret_id CHAR(26) NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    run_id CHAR(26) NOT NULL,
    folder TEXT NOT NULL DEFAULT '',
    dir TEXT NOT NULL DEFAULT '',
    job TEXT NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_secret_read_secret ON secret_read(secret_id, created_at);
CREATE INDEX IF NOT EXISTS idx_secret_read_run ON secret_read(run_id);
