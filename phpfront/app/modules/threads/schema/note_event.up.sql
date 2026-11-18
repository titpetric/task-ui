-- note_event is one item of Claude's work on a thread: a status, text,
-- thinking, a tool call and its result, the init and the result. body is
-- display text, data is compact JSON for the fields a card reads.
CREATE TABLE IF NOT EXISTS note_event (
    id CHAR(26) PRIMARY KEY NOT NULL,
    note_id CHAR(26) NOT NULL,
    run_id CHAR(26) NOT NULL DEFAULT '',
    question_id CHAR(26) NOT NULL DEFAULT '',
    seq INTEGER NOT NULL DEFAULT 0,
    kind TEXT NOT NULL CHECK (kind IN ('status', 'text', 'thinking', 'tool_use', 'tool_result', 'system', 'result', 'error')),
    tool_id TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    data TEXT NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS uidx_note_event_seq ON note_event(run_id, seq);
CREATE INDEX IF NOT EXISTS idx_note_event_thread ON note_event(note_id, question_id, run_id, seq);
