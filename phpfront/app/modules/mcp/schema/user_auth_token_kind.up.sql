-- kind tells a user's tokens apart: mcp for the ones their MCP access
-- issues, which expire and which the user page lists and deletes, and empty
-- for the ones `task-ui token` writes, which never expire and which no page
-- shows. Before it the two were told apart by whether expires_at was set,
-- which is a property of a token and not of what issued it.
ALTER TABLE user_auth_token ADD COLUMN kind TEXT NOT NULL DEFAULT '';

-- The tokens an instance already has under the one name MCP access used.
UPDATE user_auth_token SET kind = 'mcp' WHERE name = 'mcp' AND expires_at IS NOT NULL;

-- A name was unique per user, which is what held a user to a single MCP
-- token: issuing a second one replaced the first. It is now unique per user
-- and kind, so a label tells a user's MCP tokens apart and says nothing about
-- their CLI ones.
DROP INDEX IF EXISTS idx_user_auth_token_user_name;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_auth_token_user_kind_name ON user_auth_token(user_id, kind, name);
