-- shallow lists a workspace and its top-level folders alone: discovery
-- does not go deeper. Off, the default, discovery goes through the
-- checkout.
ALTER TABLE workspace ADD COLUMN shallow BOOLEAN NOT NULL DEFAULT 0;
