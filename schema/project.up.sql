-- project: What the configuration says about a folder under the root
CREATE TABLE IF NOT EXISTS project (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    branch TEXT NOT NULL DEFAULT '',
    allow TEXT NOT NULL DEFAULT '[]',
    commands TEXT NOT NULL DEFAULT '{}',
    sources TEXT NOT NULL DEFAULT '[]',
    job_allow TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_project_folder ON project(folder);
