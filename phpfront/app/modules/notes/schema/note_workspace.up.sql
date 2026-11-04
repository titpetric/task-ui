-- workspace_id is the workspace row of a prompt's folder, the default
-- workspace's for the root: a note with one is a prompt to Claude and
-- its answer, kept off the notes board. Empty for a note on the board.
ALTER TABLE note ADD COLUMN workspace_id CHAR(26) NOT NULL DEFAULT '';
-- dir is where the prompt runs, the folder and the directory inside it
-- as a slash path, or _ for the root: the page that lists its thread.
ALTER TABLE note ADD COLUMN dir TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_workspace ON note(workspace_id, dir);
