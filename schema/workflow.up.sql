-- workflow: A named chain of jobs, from any folders, run one after another
CREATE TABLE IF NOT EXISTS workflow (
    id CHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    user_id CHAR(26) NOT NULL DEFAULT '',
    user_name TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_workflow_name ON workflow(name);
