-- tool: A tool task-ui found in its PATH at start, with what it is used for
CREATE TABLE IF NOT EXISTS tool (
    name TEXT PRIMARY KEY NOT NULL,
    category TEXT NOT NULL,
    version TEXT NOT NULL DEFAULT '',
    path TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    updated_at DATETIME
);
