-- key_role: A role whose runs get an SSH key of the key folder through the run's ssh agent
CREATE TABLE IF NOT EXISTS key_role (
    id CHAR(26) PRIMARY KEY NOT NULL,
    key_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'runner', 'viewer')),
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_key_role_key_role ON key_role(key_name, role);
