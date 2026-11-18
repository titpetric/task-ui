-- Every role joins and keeps a meeting: a room is open to whoever sees it, and the creator or an admin ends and deletes one.
INSERT OR IGNORE INTO permission_policy (id, subject_kind, subject, path, effect, created_at) VALUES
    (lower(hex(randomblob(13))), 'role', 'runner', 'meeting/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'meeting/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'meeting/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'meeting/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'meeting/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'meeting/delete', 'allow', CURRENT_TIMESTAMP);
