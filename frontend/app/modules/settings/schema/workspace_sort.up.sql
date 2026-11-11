-- sort orders the workspaces in the sidebar and on the settings page,
-- lowest first; workspaces of one sort go by folder.
ALTER TABLE workspace ADD COLUMN sort INTEGER NOT NULL DEFAULT 0;
