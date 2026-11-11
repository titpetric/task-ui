-- artifact: A file a run left behind
CREATE TABLE IF NOT EXISTS artifact (
    id CHAR(26) PRIMARY KEY NOT NULL,
    run_id CHAR(26) NOT NULL,
    folder TEXT NOT NULL,
    job TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    byte_size INTEGER NOT NULL DEFAULT 0,
    content_type TEXT NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_artifact_run_id ON artifact(run_id);
CREATE INDEX IF NOT EXISTS idx_artifact_folder_job ON artifact(folder, job);
