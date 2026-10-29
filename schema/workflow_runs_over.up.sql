-- runs_over names the workspace whose projects a run of the workflow
-- walks: the steps run in each project of it, one project after
-- another. Empty, the default, runs the steps once.
ALTER TABLE workflow ADD COLUMN runs_over TEXT NOT NULL DEFAULT '';
