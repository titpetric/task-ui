-- key_read: A private key a user read, encrypted to their recipient, over MCP or the API, by its token; the last 1000 of each key kept
CREATE TABLE IF NOT EXISTS key_read (
    id CHAR(26) PRIMARY KEY NOT NULL,
    key_name TEXT NOT NULL,
    user_name TEXT NOT NULL DEFAULT '',
    token_id TEXT NOT NULL DEFAULT '',
    via TEXT NOT NULL DEFAULT 'api' CHECK (via IN ('mcp', 'api')),
    created_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_key_read_key ON key_read(key_name, created_at);
