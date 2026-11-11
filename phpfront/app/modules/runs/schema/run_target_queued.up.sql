-- state takes queued, a run of a host that waits for the host's run before
-- it, and message says why a run ended when its state needs saying, such
-- as a queued run the instance restarted before it ran. SQLite changes no
-- CHECK in place: the table moves to one with the new check and column,
-- after run_target, whose columns it keeps.
CREATE TABLE run_queued (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'running' CHECK (state IN ('queued', 'running', 'ok', 'failed', 'stopped', 'killed')),
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
    updated_at DATETIME,
    dir TEXT NOT NULL DEFAULT '',
    args TEXT NOT NULL DEFAULT '[]',
    target TEXT NOT NULL DEFAULT '',
    batch_id CHAR(26) NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT ''
);

INSERT INTO run_queued (id, folder, job, user_id, user_name, state, exit_code, forced, scheduled, term_cols, term_rows, cast_path, reason, started_at, finished_at, created_at, updated_at, dir, args, target, batch_id)
SELECT id, folder, job, user_id, user_name, state, exit_code, forced, scheduled, term_cols, term_rows, cast_path, reason, started_at, finished_at, created_at, updated_at, dir, args, target, batch_id FROM run;

DROP TABLE run;

ALTER TABLE run_queued RENAME TO run;

CREATE INDEX IF NOT EXISTS idx_run_folder_job ON run(folder, job);
CREATE INDEX IF NOT EXISTS idx_run_state ON run(state);
CREATE INDEX IF NOT EXISTS idx_run_started_at ON run(started_at);
CREATE INDEX IF NOT EXISTS idx_run_batch ON run(batch_id);
