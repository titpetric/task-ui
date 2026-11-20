-- meeting_agent adds the agent a room talks to and the note thread it answers in.
ALTER TABLE meeting ADD COLUMN agent TEXT NOT NULL DEFAULT '';
ALTER TABLE meeting ADD COLUMN agent_id CHAR(26) NOT NULL DEFAULT '';
ALTER TABLE meeting ADD COLUMN agent_note_id CHAR(26) NOT NULL DEFAULT '';
