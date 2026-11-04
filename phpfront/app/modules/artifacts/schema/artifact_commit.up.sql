-- An artifact's version: the commit of the artifact repository that holds
-- the file at path. Empty for an artifact kept as a file of its own under
-- the run's directory.
ALTER TABLE artifact ADD COLUMN commit_id TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_artifact_path ON artifact(path);
