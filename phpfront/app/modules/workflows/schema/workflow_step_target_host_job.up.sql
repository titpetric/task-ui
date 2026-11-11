-- A step of a command on a host names the host in its job, @host:<host>
-- or @host:* for every host, and keeps no target: the steps of the bare
-- @host take the job of their target.
UPDATE workflow_step SET job = '@host:' || target, target = '' WHERE job = '@host' AND target <> '';
