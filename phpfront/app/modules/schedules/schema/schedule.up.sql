-- schedule: A job that runs on its own, on a cron expression
CREATE TABLE IF NOT EXISTS schedule (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    expression TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 1,
    last_run_at DATETIME,
    next_run_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_schedule_next_run_at ON schedule(enabled, next_run_at);
CREATE INDEX IF NOT EXISTS idx_schedule_folder_job ON schedule(folder, job);
