-- refinement_comment: What a user asked for, and the run that answered it
--
-- The refinement tables came into this project from the phpscript module
-- that used to own them, where kind and collapsed_into arrived as a later
-- ALTER. They are declared here instead, in the shape the table has: a
-- database that already holds the table keeps it, because every statement
-- in these four files is IF NOT EXISTS, and ALTER TABLE ADD COLUMN is the
-- one thing SQLite cannot make conditional. Append-only starts from here.
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
    -- kind is '' for the two a person writes, the design and a correction,
    -- which the correction row tells apart, and 'apply' for the record an
    -- Apply all leaves: the prompt it sent. An apply comment is neither a
    -- design nor a rule, so the prompt is never composed out of one
    kind TEXT NOT NULL DEFAULT '',
    -- collapsed_into is the apply comment this one was folded into, and ''
    -- while it stands on its own. A folded correction is still a rule: the
    -- fold is what the page draws, and deleting is what retires a rule
    collapsed_into CHAR(26) NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refinement_comment_thread ON refinement_comment(thread_id, created_at);
