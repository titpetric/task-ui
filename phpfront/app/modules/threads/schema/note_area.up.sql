-- area marks a prompt asked from an area's page in place of a folder's:
-- workflows for the Workflows page, whose box builds and changes
-- workflows. Empty for any other note; a reply takes its thread's.
ALTER TABLE note ADD COLUMN area TEXT NOT NULL DEFAULT '';
