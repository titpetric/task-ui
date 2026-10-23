-- input is the text a run of a workflow started with, for its prompts;
-- answer is Claude's reply to the run's final prompt, or to its last
-- prompt step for a workflow without one: what the next runs' prompts
-- read of it. Empty for a run without.
ALTER TABLE workflow_run ADD COLUMN input TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow_run ADD COLUMN answer TEXT NOT NULL DEFAULT '';
