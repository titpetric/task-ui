-- note_participant: who may read a chat thread, and how far each of them has read it
CREATE TABLE IF NOT EXISTS note_participant (
    note_id CHAR(26) NOT NULL,
    user_id CHAR(26) NOT NULL,
    user_name TEXT NOT NULL DEFAULT '',
    last_read_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (note_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_note_participant_user ON note_participant(user_id);
