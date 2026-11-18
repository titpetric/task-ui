-- origin is the page or module a thread was started from: folder, workspace,
-- run, workflows, search, orb, threads, or module:<name>.
ALTER TABLE note ADD COLUMN origin TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_origin ON note(origin);
