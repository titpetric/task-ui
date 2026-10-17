-- run: One run of a job, from start to how it ended
CREATE TABLE IF NOT EXISTS run (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'running' CHECK (state IN ('running', 'ok', 'failed', 'stopped', 'killed')),
    exit_code INTEGER NOT NULL DEFAULT 0,
    forced INTEGER NOT NULL DEFAULT 0,
    scheduled INTEGER NOT NULL DEFAULT 0,
    term_cols INTEGER NOT NULL DEFAULT 120,
    term_rows INTEGER NOT NULL DEFAULT 30,
    cast_path TEXT NOT NULL DEFAULT '',
    reason TEXT NOT NULL DEFAULT '',
    started_at DATETIME,
    finished_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_run_folder_job ON run(folder, job);
CREATE INDEX IF NOT EXISTS idx_run_state ON run(state);
CREATE INDEX IF NOT EXISTS idx_run_started_at ON run(started_at);
