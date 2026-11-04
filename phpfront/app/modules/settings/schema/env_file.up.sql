-- env_file: The managed content of a dotenv file a folder reads
CREATE TABLE IF NOT EXISTS env_file (
    id CHAR(26) PRIMARY KEY NOT NULL,
    folder TEXT NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    created_at DATETIME,
    updated_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_env_file_folder_path ON env_file(folder, path);
