-- description says what a workflow is for, in a line or two, under its
-- name on its page and in the list. Empty for none.
ALTER TABLE workflow ADD COLUMN description TEXT NOT NULL DEFAULT '';
