-- meeting_participant is one user in one room: when they joined, when they were last seen, and how often they came back.
CREATE TABLE IF NOT EXISTS meeting_participant (
    meeting_id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL,
    user_name TEXT NOT NULL DEFAULT '',
    joined_at DATETIME NOT NULL,
    last_seen_at DATETIME NOT NULL,
    left_at DATETIME,
    joins INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (meeting_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_meeting_participant_meeting ON meeting_participant(meeting_id, left_at);
