-- note: A message on the notes board or in the chat, or a reply to one
CREATE TABLE IF NOT EXISTS note (
    id CHAR(26) PRIMARY KEY NOT NULL,
    channel TEXT NOT NULL DEFAULT 'notes' CHECK (channel IN ('notes', 'chat')),
    parent_id CHAR(26) NOT NULL DEFAULT '',
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL,
    body TEXT NOT NULL,
    -- The users the body names with @, space separated, between spaces.
    mentions TEXT NOT NULL DEFAULT '',
    -- How the instruction to the agent goes: empty for a note to nobody,
    -- open, working, done or failed.
    state TEXT NOT NULL DEFAULT '' CHECK (state IN ('', 'open', 'working', 'done', 'failed')),
    -- The run the agent did the work in.
    run_id CHAR(26) NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_note_channel ON note(channel, parent_id);
CREATE INDEX IF NOT EXISTS idx_note_parent ON note(parent_id);
CREATE INDEX IF NOT EXISTS idx_note_state ON note(state);
