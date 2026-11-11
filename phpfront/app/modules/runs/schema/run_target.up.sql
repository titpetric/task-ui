-- target is the host a run of @host ran its command on, from the
-- inventory; empty for every other job.
ALTER TABLE run ADD COLUMN target TEXT NOT NULL DEFAULT '';
-- batch_id groups the runs started together, the host runs of one Run
-- on the inventory: the ID of the batch's first run, its own included.
-- Empty for a run started alone.
ALTER TABLE run ADD COLUMN batch_id CHAR(26) NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_run_batch ON run(batch_id);
