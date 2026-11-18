-- meeting is one room the orb keeps: who opened it, when, and what became of the transcript.
CREATE TABLE IF NOT EXISTS meeting (
    id CHAR(26) PRIMARY KEY NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    user_id CHAR(26) NOT NULL,
    user_name TEXT NOT NULL DEFAULT '',
    state TEXT NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'ended')),
    started_at DATETIME NOT NULL,
    ended_at DATETIME,
    limit_seconds INTEGER NOT NULL DEFAULT 3600,
    empty_since DATETIME,
    summary TEXT NOT NULL DEFAULT '',
    summary_state TEXT NOT NULL DEFAULT '' CHECK (summary_state IN ('', 'working', 'done', 'failed')),
    summary_message TEXT NOT NULL DEFAULT '',
    summary_user_name TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_meeting_state ON meeting(state, started_at);
