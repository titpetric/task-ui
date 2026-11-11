-- A step asks Claude in place of a job: prompt is what it asks, with the
-- output of the steps before it, folder is where Claude runs, a folder
-- or _ for the root, dir a directory inside it, and job is
-- @claude-prompt. Empty for a step of a job or a workflow. A run's step
-- keeps the prompt it asked and Claude's answer, the step's output for
-- the prompts after it.
ALTER TABLE workflow_step ADD COLUMN prompt TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow_step ADD COLUMN dir TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow_run_step ADD COLUMN prompt TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow_run_step ADD COLUMN answer TEXT NOT NULL DEFAULT '';
