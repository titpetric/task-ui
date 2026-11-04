-- refinement_thread: A design being refined, and the report that indexes its artifacts
CREATE TABLE IF NOT EXISTS refinement_thread (
    id CHAR(26) PRIMARY KEY NOT NULL,
    -- title is the first line of the comment that started it, for the list
    title TEXT NOT NULL DEFAULT '',
    -- user is who opened it, by name
    user TEXT NOT NULL DEFAULT '',
    -- report_markdown is the state of README.md in the artifacts: the index
    -- a run writes of everything the thread's comments produced. The runs
    -- write the file, and the thread reads it back here so the page draws
    -- the index without reading the artifact tree on every request
    report_markdown TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refinement_thread_created ON refinement_thread(created_at);
