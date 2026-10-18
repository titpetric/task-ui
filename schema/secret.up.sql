-- secret: A variable added to the environment of runs, for every folder or for one
CREATE TABLE IF NOT EXISTS secret (
    id CHAR(26) PRIMARY KEY NOT NULL,
    scope TEXT NOT NULL,
    name TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_secret_scope_name ON secret(scope, name);
