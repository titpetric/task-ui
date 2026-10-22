-- shallow lists a project and its top-level folders alone: discovery
-- does not go deeper. Off, the default, discovery goes through the
-- checkout.
ALTER TABLE project RENAME COLUMN root_only TO shallow;
