-- pin: what a user pins to the sidebar, a folder (a workspace, a project,
-- a folder or a directory inside one) or a job of it, the user's own
CREATE TABLE IF NOT EXISTS pin (
    id CHAR(26) PRIMARY KEY NOT NULL,
    user_name TEXT NOT NULL,
    -- folder is the folder's ID, or a directory's inside one
    folder TEXT NOT NULL,
    -- job is the job a job pin runs; empty pins the folder
    job TEXT NOT NULL DEFAULT '',
    -- dir is the directory inside the folder the job runs in, a slash path
    dir TEXT NOT NULL DEFAULT '',
    -- args are the arguments the job runs with after --, a JSON list
    args TEXT NOT NULL DEFAULT '[]',
    created_at DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS pin_user_target ON pin (user_name, folder, job, dir, args);
