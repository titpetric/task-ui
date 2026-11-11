-- host: A machine of the inventory, entered by a person or taken from the Beszel hub; nothing but a person takes one out
CREATE TABLE IF NOT EXISTS host (
    id CHAR(26) PRIMARY KEY NOT NULL,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    source TEXT NOT NULL CHECK (source IN ('manual', 'beszel')),
    login TEXT NOT NULL DEFAULT '',
    added_by TEXT NOT NULL DEFAULT '',
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_host_address ON host(address);
