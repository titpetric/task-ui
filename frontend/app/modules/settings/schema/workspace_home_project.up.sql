-- home_project_id is the project of the workspace whose directory its
-- runs get as HOME, '' for none: the server's HOME. A workspace is a
-- plain directory of sibling checkouts, its projects, so the repository
-- and the branch are the projects' and leave the workspace's row.
ALTER TABLE workspace ADD COLUMN home_project_id CHAR(26) NOT NULL DEFAULT '';
ALTER TABLE workspace DROP COLUMN source_url;
ALTER TABLE workspace DROP COLUMN branch;
