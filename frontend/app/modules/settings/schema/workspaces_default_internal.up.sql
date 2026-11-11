-- The default workspace, the root, is named internal and no longer _:
-- every stored reference to _ moves to internal. Rows older runs of the
-- Claude jobs left under internal stay, and now sit in the root.
UPDATE workspace SET folder = 'internal' WHERE folder = '_';
UPDATE project SET folder = 'internal' WHERE folder = '_';
UPDATE run SET folder = 'internal' WHERE folder = '_';
UPDATE schedule SET folder = 'internal' WHERE folder = '_';
UPDATE artifact SET folder = 'internal' WHERE folder = '_';
UPDATE pin SET folder = 'internal' WHERE folder = '_';
UPDATE env_file SET folder = 'internal' WHERE folder = '_';
UPDATE secret_read SET folder = 'internal' WHERE folder = '_';
UPDATE workflow_step SET folder = 'internal' WHERE folder = '_';
UPDATE workflow_run_step SET folder = 'internal' WHERE folder = '_';
UPDATE workflow SET prompt_folder = 'internal' WHERE prompt_folder = '_';
UPDATE secret SET scope = 'workspace:internal' WHERE scope = 'workspace:_';
UPDATE key_info SET scope = 'workspace:internal' WHERE scope = 'workspace:_';
UPDATE note SET workspace_id = 'internal' WHERE workspace_id = '_';
UPDATE note SET dir = 'internal' WHERE dir = '_';
UPDATE workspace_key_grant SET workspace_id = 'internal' WHERE workspace_id = '_';
