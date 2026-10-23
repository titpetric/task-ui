-- next_run_at and last_run_at hold UTC, as Go writes a time: the due
-- check compares them as text against the clock in UTC. A row written
-- with the local offset, 2026-10-23 07:00:00 +0200 CEST, becomes the same
-- instant in UTC, 2026-10-23 05:00:00 +0000 UTC.
UPDATE schedule SET next_run_at = strftime('%Y-%m-%d %H:%M:%S', substr(next_run_at, 1, 19),
	printf('%d minutes', (CASE substr(substr(next_run_at, 20), instr(substr(next_run_at, 20), ' ') + 1, 1) WHEN '-' THEN 1 ELSE -1 END)
		* (CAST(substr(substr(next_run_at, 20), instr(substr(next_run_at, 20), ' ') + 2, 2) AS INTEGER) * 60
		+ CAST(substr(substr(next_run_at, 20), instr(substr(next_run_at, 20), ' ') + 4, 2) AS INTEGER)))) || ' +0000 UTC'
WHERE next_run_at IS NOT NULL AND substr(substr(next_run_at, 20), instr(substr(next_run_at, 20), ' ') + 1, 5) <> '+0000';

UPDATE schedule SET last_run_at = strftime('%Y-%m-%d %H:%M:%S', substr(last_run_at, 1, 19),
	printf('%d minutes', (CASE substr(substr(last_run_at, 20), instr(substr(last_run_at, 20), ' ') + 1, 1) WHEN '-' THEN 1 ELSE -1 END)
		* (CAST(substr(substr(last_run_at, 20), instr(substr(last_run_at, 20), ' ') + 2, 2) AS INTEGER) * 60
		+ CAST(substr(substr(last_run_at, 20), instr(substr(last_run_at, 20), ' ') + 4, 2) AS INTEGER)))) || ' +0000 UTC'
WHERE last_run_at IS NOT NULL AND substr(substr(last_run_at, 20), instr(substr(last_run_at, 20), ' ') + 1, 5) <> '+0000';
