-- sidebar_hidden keeps a workflow out of the sidebar, for one that is a
-- step of others rather than a button of its own. Off, the default, it
-- is listed.
ALTER TABLE workflow ADD COLUMN sidebar_hidden BOOLEAN NOT NULL DEFAULT 0;
