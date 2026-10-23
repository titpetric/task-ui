-- The runner role writes the artifact tree, as an agent with a runner's
-- token does over MCP: artifact/edit writes a file of the tree, and takes
-- one out of it.
INSERT OR IGNORE INTO permission_policy (id, subject_kind, subject, path, effect, created_at) VALUES
    (lower(hex(randomblob(13))), 'role', 'runner', 'artifact/edit', 'allow', CURRENT_TIMESTAMP);
