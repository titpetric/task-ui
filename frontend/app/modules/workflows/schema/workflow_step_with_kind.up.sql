-- kind says what a step does: job runs a job of a folder, workflow runs
-- another workflow, prompt asks Claude its prompt in its folder and dir,
-- and summary asks Claude task-ui's summary of the steps before it,
-- with no prompt of its own. A run's step keeps the kind it ran as.
ALTER TABLE workflow_step ADD COLUMN kind TEXT NOT NULL DEFAULT 'job' CHECK (kind IN ('job', 'workflow', 'prompt', 'summary'));
ALTER TABLE workflow_run_step ADD COLUMN kind TEXT NOT NULL DEFAULT 'job' CHECK (kind IN ('job', 'workflow', 'prompt', 'summary'));
UPDATE workflow_step SET kind = 'workflow' WHERE workflow_ref <> '';
UPDATE workflow_step SET kind = 'prompt' WHERE prompt <> '';
UPDATE workflow_run_step SET kind = 'workflow' WHERE workflow_ref <> '';
UPDATE workflow_run_step SET kind = 'prompt' WHERE prompt <> '';
-- A workflow's final prompt is its last prompt step, in the place
-- prompt_folder names, the root when empty; the workflow's prompt and
-- prompt_folder stay empty.
INSERT INTO workflow_step (id, workflow_id, position, folder, job, kind, prompt, dir)
SELECT lower(hex(randomblob(13))), w.id,
    COALESCE((SELECT MAX(s.position) + 1 FROM workflow_step s WHERE s.workflow_id = w.id), 0),
    CASE WHEN w.prompt_folder = '' THEN '_' ELSE w.prompt_folder END,
    '@claude-prompt', 'prompt', trim(w.prompt), ''
FROM workflow w WHERE trim(w.prompt) <> '';
UPDATE workflow SET prompt = '', prompt_folder = '' WHERE prompt <> '' OR prompt_folder <> '';
