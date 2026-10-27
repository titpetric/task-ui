-- secret: a variable added to runs is either a secret, whose value no page
-- shows, or an env entry, whose value the page shows. Both are encrypted at
-- rest and reach a run's environment; task-ui get-secret reads either.
ALTER TABLE secret ADD COLUMN is_secret INTEGER NOT NULL DEFAULT 1;
