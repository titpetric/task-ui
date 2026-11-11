-- sidebar_hidden keeps a project, and the folders inside it, out of the
-- sidebar; its pages stay reachable. Off, the default, it is listed.
ALTER TABLE project ADD COLUMN sidebar_hidden BOOLEAN NOT NULL DEFAULT 0;
