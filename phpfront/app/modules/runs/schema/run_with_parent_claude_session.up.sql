-- claude_session is the ULID of the Claude Code conversation a prompt's run
-- carries, which names its directory under data/claude and resumes it after
-- a restart; empty for every other job.
ALTER TABLE run ADD COLUMN claude_session CHAR(26) NOT NULL DEFAULT '';
