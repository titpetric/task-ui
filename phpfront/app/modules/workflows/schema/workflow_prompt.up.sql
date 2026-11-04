-- prompt is what a workflow asks Claude once its steps end, as the user
-- who started the run: a summary, or work on the steps' output, or the
-- whole workflow when it has no step to run. prompt_folder is where
-- Claude runs, a folder or folder/dir; empty runs it in the root, _.
ALTER TABLE workflow ADD COLUMN prompt TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow ADD COLUMN prompt_folder TEXT NOT NULL DEFAULT '';
