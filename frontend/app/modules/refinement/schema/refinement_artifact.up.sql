-- refinement_artifact: A file a thread's run left in the instance's artifact tree
CREATE TABLE IF NOT EXISTS refinement_artifact (
    id CHAR(26) PRIMARY KEY NOT NULL,
    thread_id CHAR(26) NOT NULL,
    -- comment_id is the comment whose run wrote it
    comment_id CHAR(26) NOT NULL DEFAULT '',
    -- path is where it sits in the artifact tree, which is what
    -- /artifacts/files/<path> serves and what the report links
    path TEXT NOT NULL,
    -- kind is 'image' for a file the page embeds, 'file' for one it links.
    -- A schema diagram is an image and the page draws it rather than
    -- offering a download
    kind TEXT NOT NULL DEFAULT 'file',
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_refinement_artifact_thread ON refinement_artifact(thread_id, path);
