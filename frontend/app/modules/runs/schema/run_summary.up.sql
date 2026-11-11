-- run_summary: Claude's summary of a finished run's output, kept so the run is summarized once
CREATE TABLE IF NOT EXISTS run_summary (
    run_id CHAR(26) PRIMARY KEY NOT NULL,
    summary TEXT NOT NULL,
    user_name TEXT NOT NULL,
    created_at DATETIME
);
