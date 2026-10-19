-- ssh_key_policy: A role whose runs get an instance SSH key through the run's ssh agent
CREATE TABLE IF NOT EXISTS ssh_key_policy (
    id CHAR(26) PRIMARY KEY NOT NULL,
    key_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'runner', 'viewer')),
    created_at DATETIME
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ssh_key_policy_key_role ON ssh_key_policy(key_name, role);
CREATE INDEX IF NOT EXISTS idx_ssh_key_policy_role ON ssh_key_policy(role);
