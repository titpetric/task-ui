-- sidebar_limit is how many of the workspace's projects the sidebar
-- lists, the most recently committed first; 0, the default, lists every
-- one, by name.
ALTER TABLE workspace ADD COLUMN sidebar_limit INTEGER NOT NULL DEFAULT 0;
