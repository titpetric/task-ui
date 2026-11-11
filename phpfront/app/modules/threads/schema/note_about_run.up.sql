-- about_run_id is the run a prompt asks about, from that run's page: its
-- output and how it ended go to Claude with the thread, and the thread
-- lists on the run's page, not with its directory's prompts. Empty for
-- every other note. run_id stays the run that answers.
ALTER TABLE note ADD COLUMN about_run_id CHAR(26) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_about_run ON note(about_run_id);
