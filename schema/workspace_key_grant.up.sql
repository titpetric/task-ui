-- workspace_key_grant: A role or a user, by name, whose runs in a workspace get one of its SSH keys; a key with no grant goes to every run in the workspace
CREATE TABLE IF NOT EXISTS workspace_key_grant (
    id CHAR(26) PRIMARY KEY NOT NULL,
    workspace_id CHAR(26) NOT NULL,
    key_name TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    CHECK ((role = '') <> (user_name = ''))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workspace_key_grant_key ON workspace_key_grant(workspace_id, key_name, role, user_name);
CREATE INDEX IF NOT EXISTS idx_workspace_key_grant_workspace_id ON workspace_key_grant(workspace_id);
