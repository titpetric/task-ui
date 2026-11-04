-- run_as_user_id is the user every run of the workflow runs as, its
-- steps with that user's permissions, Claude token and workspace keys;
-- empty, the default, is the workflow's owner.
ALTER TABLE workflow ADD COLUMN run_as_user_id TEXT NOT NULL DEFAULT '';
