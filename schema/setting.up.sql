-- setting: An instance setting a page edits, by name, such as the data retention rules
CREATE TABLE IF NOT EXISTS setting (
    name TEXT PRIMARY KEY NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    updated_at DATETIME
);
