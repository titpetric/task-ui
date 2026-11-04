-- home_url and home_branch name the repository of the workspace's home,
-- the checkout at <workspace>/home that its runs get as HOME, '' for an
-- empty home. The home is a folder of its own and no project, so the row
-- of the project that was the home goes, and home_project_id stays
-- unused.
ALTER TABLE workspace ADD COLUMN home_url TEXT NOT NULL DEFAULT '';
ALTER TABLE workspace ADD COLUMN home_branch TEXT NOT NULL DEFAULT '';
UPDATE workspace SET
    home_url = COALESCE((SELECT p.source_url FROM project p WHERE p.id = workspace.home_project_id), ''),
    home_branch = COALESCE((SELECT p.branch FROM project p WHERE p.id = workspace.home_project_id), '')
WHERE home_project_id <> '';
DELETE FROM project WHERE id IN (SELECT home_project_id FROM workspace WHERE home_project_id <> '');
UPDATE workspace SET home_project_id = '';
