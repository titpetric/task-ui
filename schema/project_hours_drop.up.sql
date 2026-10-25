-- The hours of a project and of its jobs go: the workspace holds the one
-- hours rule its folders run under.
ALTER TABLE project DROP COLUMN hours;
ALTER TABLE project DROP COLUMN job_hours;
