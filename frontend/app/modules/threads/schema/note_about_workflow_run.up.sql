-- about_workflow_run_id is the workflow run whose prompt a thread asks,
-- once the run's steps end: the steps' output goes to Claude with the
-- thread, and the thread lists on the workflow run's page, not with its
-- directory's prompts. Empty for every other note.
ALTER TABLE note ADD COLUMN about_workflow_run_id CHAR(26) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_about_workflow_run ON note(about_workflow_run_id);
