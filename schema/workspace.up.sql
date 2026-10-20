-- workspace: A checkout under the root that holds projects: its folder
-- is the namespace of the projects cloned inside it, and any number of
-- workspaces may share a repository.
CREATE TABLE IF NOT EXISTS workspace (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_folder ON workspace(folder);
