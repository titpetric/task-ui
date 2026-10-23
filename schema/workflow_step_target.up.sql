-- target and args are what a step of @host runs: the host, or * for
-- every host with an ssh login, and the command, a JSON list. Empty for
-- a step of any other job.
ALTER TABLE workflow_step ADD COLUMN target TEXT NOT NULL DEFAULT '';
ALTER TABLE workflow_step ADD COLUMN args TEXT NOT NULL DEFAULT '[]';
