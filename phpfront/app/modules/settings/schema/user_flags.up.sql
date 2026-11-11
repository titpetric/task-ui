-- admin gives a user every permission, whatever the role column says;
-- role is the role without it, runner or viewer. mcp lets a user's token
-- reach the MCP endpoint, and agent_reply has a note that names the
-- user start the Notes agent as them. The admins of the role lists
-- become admin runners, and a user holding a token keeps MCP.
ALTER TABLE user ADD COLUMN admin BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE user ADD COLUMN mcp BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE user ADD COLUMN agent_reply BOOLEAN NOT NULL DEFAULT 0;

UPDATE user SET admin = 1, role = 'runner' WHERE role = 'admin';
UPDATE user SET mcp = 1 WHERE id IN (SELECT user_id FROM user_auth_token);
