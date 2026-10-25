-- workflow_id names the workflow a schedule of a workflow starts, a
-- run of every step as the schedule's user; such a schedule leaves
-- folder and job empty. Empty for a schedule of a job.
ALTER TABLE schedule ADD COLUMN workflow_id CHAR(26) NOT NULL DEFAULT '';
