-- meeting_line is one entry of a room's log: what a user said, or a join, a leave, a rejoin or the end.
CREATE TABLE IF NOT EXISTS meeting_line (
    id CHAR(26) PRIMARY KEY NOT NULL,
    meeting_id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    kind TEXT NOT NULL DEFAULT 'said' CHECK (kind IN ('said', 'join', 'leave', 'rejoin', 'end')),
    body TEXT NOT NULL DEFAULT '',
    start_ms INTEGER NOT NULL DEFAULT 0,
    end_ms INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_meeting_line_meeting ON meeting_line(meeting_id, created_at);
