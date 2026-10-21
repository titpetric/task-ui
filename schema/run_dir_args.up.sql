-- dir is the directory inside the folder the run's job ran in, a slash
-- path; empty for a job of the folder itself.
ALTER TABLE run ADD COLUMN dir TEXT NOT NULL DEFAULT '';
-- args are the arguments the run passed its job after --, a JSON list.
ALTER TABLE run ADD COLUMN args TEXT NOT NULL DEFAULT '[]';
