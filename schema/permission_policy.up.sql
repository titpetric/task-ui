-- permission_policy: A role's or a user's entry for a rule of permissions.yml: allow or deny; no entry inherits, and a rule nobody allows is denied
CREATE TABLE IF NOT EXISTS permission_policy (
    id CHAR(26) PRIMARY KEY NOT NULL,
    subject_kind TEXT NOT NULL CHECK (subject_kind IN ('role', 'user')),
    subject TEXT NOT NULL,
    path TEXT NOT NULL,
    effect TEXT NOT NULL CHECK (effect IN ('allow', 'deny')),
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_permission_policy_subject_path ON permission_policy(subject_kind, subject, path);

-- The roles start with what they could do before there were policies:
-- a runner runs jobs, and both read and write notes, workflows and
-- schedules of their own; the admin role allows everything and holds no
-- entries.
INSERT OR IGNORE INTO permission_policy (id, subject_kind, subject, path, effect, created_at) VALUES
    (lower(hex(randomblob(13))), 'role', 'runner', 'run', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'note/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'note/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'note/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'workflow/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'workflow/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'workflow/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'schedule/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'schedule/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'schedule/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'artifact/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'runner', 'inventory/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'note/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'note/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'note/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'workflow/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'workflow/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'workflow/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'schedule/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'schedule/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'schedule/delete', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'artifact/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'role', 'viewer', 'inventory/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'user', 'claude', 'note/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'user', 'claude', 'note/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'user', 'claude', 'prompt/view', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'user', 'claude', 'prompt/edit', 'allow', CURRENT_TIMESTAMP),
    (lower(hex(randomblob(13))), 'user', 'claude', 'run', 'allow', CURRENT_TIMESTAMP);
