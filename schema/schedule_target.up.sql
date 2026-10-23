-- target and args are what a schedule of @host runs: the host, or * for
-- every host with an ssh login, and the command, a JSON list. Empty for
-- a schedule of any other job.
ALTER TABLE schedule ADD COLUMN target TEXT NOT NULL DEFAULT '';
ALTER TABLE schedule ADD COLUMN args TEXT NOT NULL DEFAULT '[]';
