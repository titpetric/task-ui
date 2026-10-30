-- refinement_comment: What a user asked for, and the run that answered it
CREATE TABLE IF NOT EXISTS refinement_comment (
    id CHAR(26) PRIMARY KEY NOT NULL,
    -- thread_id is the thread it belongs to; no foreign key, as everywhere
    -- in this schema
    thread_id CHAR(26) NOT NULL,
    -- user is who wrote it, by name
    user TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    -- note_id is the thread Task::dispatch started, which holds the run and
    -- Claude's reply. Empty while the dispatch has not returned
    note_id CHAR(26) NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refinement_comment_thread ON refinement_comment(thread_id, created_at);
