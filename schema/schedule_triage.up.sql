-- triage says a failed run of the schedule gets Claude's triage: a
-- prompt thread on the run's page that reads the run, tries to fix the
-- cause in the folder and ends with a verdict, and a note on the board.
-- Off by default; the schedule's user turns it on.
ALTER TABLE schedule ADD COLUMN triage INTEGER NOT NULL DEFAULT 0;
