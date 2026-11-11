-- refinement_correction: A comment that changed what an earlier run produced
CREATE TABLE IF NOT EXISTS refinement_correction (
    id CHAR(26) PRIMARY KEY NOT NULL,
    thread_id CHAR(26) NOT NULL,
    -- comment_id is the comment that asked for the change; corrects_id is
    -- the comment whose artifacts it changed, the one before it
    comment_id CHAR(26) NOT NULL,
    corrects_id CHAR(26) NOT NULL DEFAULT '',
    -- A thread's first comment is no correction: it has nothing to correct.
    -- Every comment after it is one, which is what separates "design this"
    -- from "use ULID as primary keys"
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refinement_correction_thread ON refinement_correction(thread_id, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS uidx_refinement_correction_comment ON refinement_correction(comment_id);
