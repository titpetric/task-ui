-- user_name is who wrote a file of the artifact tree over the API or
-- MCP, with no run: such an artifact has no run, folder or job. Empty
-- for a file a run left.
ALTER TABLE artifact ADD COLUMN user_name TEXT NOT NULL DEFAULT '';
