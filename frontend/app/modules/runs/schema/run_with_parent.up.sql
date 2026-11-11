-- parent_id is the ID of the run or workflow run that dispatched this
-- run; the sidebar groups the run under that one instead of listing it
-- on its own. Empty for a run somebody started. The name sorts after
-- run_target_queued, which rebuilds the table from a column list a new
-- column has to follow.
ALTER TABLE run ADD COLUMN parent_id CHAR(26) NOT NULL DEFAULT '';
-- parent_kind says what parent_id names, run or workflow_run; empty with
-- an empty parent_id.
ALTER TABLE run ADD COLUMN parent_kind TEXT NOT NULL DEFAULT '';

-- The order is the listing's, started_at then id: on parent_id alone the
-- roots query matches almost every row and sorts them all, which is
-- worse than the scan of idx_run_started_at it takes the place of.
CREATE INDEX IF NOT EXISTS idx_run_parent ON run(parent_id, started_at, id);
