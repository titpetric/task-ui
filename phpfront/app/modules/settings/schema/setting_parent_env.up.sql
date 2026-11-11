-- An instance that has runs already keeps passing the whole server
-- environment to them, as before the parent environment setting; a
-- fresh one starts on none.
INSERT OR IGNORE INTO setting (name, value, updated_at)
SELECT 'runs.parent_env', 'all', NULL
WHERE EXISTS (SELECT 1 FROM run);
