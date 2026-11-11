-- edited_at is when its author, or an admin, last changed the body of a
-- note: empty for a note as posted. updated_at moves with the state too.
ALTER TABLE note ADD COLUMN edited_at DATETIME;
