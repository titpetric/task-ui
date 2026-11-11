-- expires_at is when a token stops signing in: an MCP token made from a
-- user's settings lasts 30 days. Empty never expires, as the tokens of
-- `task-ui token` do.
ALTER TABLE user_auth_token ADD COLUMN expires_at DATETIME;
