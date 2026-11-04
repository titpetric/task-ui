-- A command on a host has a job a host, @host:<host>: the runs of the
-- bare @host take the job of their target.
UPDATE run SET job = '@host:' || target WHERE job = '@host' AND target <> '';
