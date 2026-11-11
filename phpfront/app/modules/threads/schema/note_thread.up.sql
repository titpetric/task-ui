-- deleted_at is when the thread was deleted, and a deleted note takes its
-- replies out of every listing with it. The rows stay: a thread is not
-- edited, so the record of what was said is what the delete keeps.
ALTER TABLE note ADD COLUMN deleted_at DATETIME;

-- audio_path is the artifact the message was spoken into, or the one the
-- answer was read out as, relative to the artifact tree. Empty for a
-- message nobody spoke.
ALTER TABLE note ADD COLUMN audio_path TEXT NOT NULL DEFAULT '';

-- spoken is the transcription as it was said, without the instructions the
-- body carries for the agent: what the thread shows under the player.
-- Empty for a message that was typed, whose body is what it says.
ALTER TABLE note ADD COLUMN spoken TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_note_deleted ON note(deleted_at);
