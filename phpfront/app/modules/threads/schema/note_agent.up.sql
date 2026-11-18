-- agent is the user a thread answers as, from the users with agent_reply
-- and mcp: claude by default, or the agent picked for it.
ALTER TABLE note ADD COLUMN agent TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_agent ON note(agent);
