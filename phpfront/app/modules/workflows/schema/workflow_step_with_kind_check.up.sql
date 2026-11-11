-- kind takes check, a step that reads the step before it and sends the
-- workflow's steps round again when it finds one of its markers there:
-- markers is what it looks for, a comma separated list, and tries the most
-- times the steps run, counting the first, 0 for the default. A run's step
-- keeps the markers it looked for, and says in its answer what it found.
-- SQLite changes no CHECK in place: each table moves to one with the new
-- check and columns, after workflow_step_with_kind, whose columns it keeps.
CREATE TABLE workflow_step_check (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workflow_id CHAR(26) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    disabled BOOLEAN NOT NULL DEFAULT 0,
    prompt TEXT NOT NULL DEFAULT '',
    dir TEXT NOT NULL DEFAULT '',
    workflow_ref CHAR(26) NOT NULL DEFAULT '',
    target TEXT NOT NULL DEFAULT '',
    args TEXT NOT NULL DEFAULT '[]',
    kind TEXT NOT NULL DEFAULT 'job' CHECK (kind IN ('job', 'workflow', 'prompt', 'summary', 'check')),
    markers TEXT NOT NULL DEFAULT '',
    tries INTEGER NOT NULL DEFAULT 0
);

INSERT INTO workflow_step_check (id, workflow_id, position, folder, job, disabled, prompt, dir, workflow_ref, target, args, kind)
SELECT id, workflow_id, position, folder, job, disabled, prompt, dir, workflow_ref, target, args, kind FROM workflow_step;

DROP TABLE workflow_step;

ALTER TABLE workflow_step_check RENAME TO workflow_step;

CREATE INDEX IF NOT EXISTS idx_workflow_step_workflow ON workflow_step(workflow_id, position);

CREATE TABLE workflow_run_step_check (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workflow_run_id CHAR(26) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    run_id CHAR(26) NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'running', 'ok', 'failed', 'stopped', 'killed', 'skipped')),
    created_at DATETIME,
    updated_at DATETIME,
    prompt TEXT NOT NULL DEFAULT '',
    answer TEXT NOT NULL DEFAULT '',
    workflow_ref CHAR(26) NOT NULL DEFAULT '',
    workflow_run_ref CHAR(26) NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'job' CHECK (kind IN ('job', 'workflow', 'prompt', 'summary', 'check')),
    markers TEXT NOT NULL DEFAULT ''
);

INSERT INTO workflow_run_step_check (id, workflow_run_id, position, folder, job, run_id, state, created_at, updated_at, prompt, answer, workflow_ref, workflow_run_ref, kind)
SELECT id, workflow_run_id, position, folder, job, run_id, state, created_at, updated_at, prompt, answer, workflow_ref, workflow_run_ref, kind FROM workflow_run_step;

DROP TABLE workflow_run_step;

ALTER TABLE workflow_run_step_check RENAME TO workflow_run_step;

CREATE INDEX IF NOT EXISTS idx_workflow_run_step_run ON workflow_run_step(workflow_run_id, position);
